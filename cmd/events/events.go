package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/signal"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/alerts"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/amqp"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/api"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/cacheutil"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/config"
	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/elasticsearch"

	"flag"
	"fmt"
	"log"
	"os"
)

/*************** main routine ***********************/
// eventusage and command-line flags
func eventusage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run events/main.go -config sa.events.config.json -debug
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200
	**************************************************************
	For running with AMQP ,Prometheus,API and AlertManager use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200 -alertmanager=http://localhost:9090/v1/api/alert -apiurl=localhost:8082 -amqppublishurl=127.0.0.1:5672/collectd/alert
	**************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

var (
	debuge          = func(format string, data ...interface{}) {} // Default no debugging output
	amqpEventServer *amqp10.AMQPServer
	amqpHandler     *amqp10.AMQPHandler
	serverConfig    saconfig.EventConfiguration
	elasticClient   *saelastic.ElasticClient
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// set flags for parsing options
	flag.Usage = eventusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fAMQP1EventURL := flag.String("amqp1EventURL", "", "AMQP1.0 events listener example 127.0.0.1:5672/collectd/notify")
	fElasticHostURL := flag.String("eshost", "", "ElasticSearch host http://localhost:9200")
	fAlertManagerURL := flag.String("alertmanager", "", "(Optional)AlertManager endpoint http://localhost:9090/v1/api/alert")
	fAPIEndpointURL := flag.String("apiurl", "", "(Optional)API endpoint localhost:8082")
	fAMQP1PublishURL := flag.String("amqppublishurl", "", "(Optional) AMQP1.0 event publish address 127.0.0.1:5672/collectd/alert")
	fResetIndex := flag.Bool("resetIndex", false, "Optional Clean all index before on start (default false)")
	fPrefetch := flag.Int("prefetch", 0, "AMQP1.0 option: Enable prefetc and set capacity(0 is disabled,>0 enabled with capacity of >0) (OPTIONAL)")
	fUniqueName := flag.String("uname", "metrics-"+strconv.Itoa(rand.Intn(100)), "Unique name across application")

	flag.Parse()

	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadEventConfig(*fConfigLocation)
		if *fDebug {
			serverConfig.Debug = true
		}
	} else {
		serverConfig = saconfig.EventConfiguration{
			AMQP1EventURL:   *fAMQP1EventURL,
			ElasticHostURL:  *fElasticHostURL,
			AlertManagerURL: *fAlertManagerURL,
			Prefetch:        *fPrefetch,
			API: saconfig.EventAPIConfig{
				APIEndpointURL:  *fAPIEndpointURL,
				AMQP1PublishURL: *fAMQP1PublishURL,
			},
			ResetIndex: *fResetIndex,
			Debug:      *fDebug,
			UniqueName: *fUniqueName,
		}

	}
	if serverConfig.Debug {
		debuge = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if len(serverConfig.AMQP1EventURL) == 0 {
		log.Println("AMQP1 Event URL is required")
		eventusage()
		os.Exit(1)
	}
	if len(serverConfig.ElasticHostURL) == 0 {
		log.Println("Elastic Host URL is required")
		eventusage()
		os.Exit(1)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("caught sig: %+v", sig)
			log.Println("Wait for 2 second to finish processing")
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
	}()

	if len(serverConfig.AlertManagerURL) > 0 {
		log.Printf("AlertManager configured at %s\n", serverConfig.AlertManagerURL)
		serverConfig.AlertManagerEnabled = true
	} else {
		log.Println("AlertManager disabled")
	}
	if len(serverConfig.API.APIEndpointURL) > 0 {
		log.Printf("API available at %s\n", serverConfig.API.APIEndpointURL)
		serverConfig.APIEnabled = true
	} else {
		log.Println("API disabled")
	}
	if len(serverConfig.API.AMQP1PublishURL) > 0 {
		log.Printf("AMQP1.0 Publish address at %s\n", serverConfig.API.AMQP1PublishURL)
		serverConfig.PublishEventEnabled = true
	} else {
		log.Println("AMQP1.0 Publish address disabled")
	}

	/* Print Configuration details */
	//mertic handler for event mertics to check health status
	applicationHealth := cacheutil.NewApplicationHealthCache()
	metricHandler := apihandler.NewAppStateEventMetricHandler(applicationHealth)
	amqpHandler := amqp10.NewAMQPHandler("Event Consumer")
	debuge("Debug:Config %#v\n", serverConfig)

	///Metric Listener
	amqpEventsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1EventURL)
	log.Printf("Connecting to AMQP1 : %s\n", amqpEventsurl)
	// done channel is used during testing
	done := make(chan bool)
	amqpEventServer = amqp10.NewAMQPServer(amqpEventsurl, serverConfig.Debug, -1, serverConfig.Prefetch, amqpHandler, done, false, *fUniqueName)

	log.Printf("Listening.....\n")

	log.Printf("Connecting to ElasticSearch : %s\n", serverConfig.ElasticHostURL)
	elasticClient = saelastic.CreateClient(serverConfig.ElasticHostURL, serverConfig.ResetIndex, serverConfig.Debug)
	applicationHealth.ElasticSearchState = 1

	/**** HTTP Listener for alerts from alert manager *******************************
	*
	*
	********************************************************************************/
	//configure http alert route to amqp1.0
	if serverConfig.APIEnabled {
		prometheus.MustRegister(metricHandler, amqpHandler)
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())

		context := apihandler.NewAPIContext(serverConfig)
		http.Handle("/alert", apihandler.Handler{context, apihandler.AlertHandler}) //creates writer everytime api is called.
		http.Handle("/metrics", prometheus.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
																	<head><title>Smart Gateway Event API</title></head>
																	<body>
																	<h1>APi </h1>
																	/alerts Post alerts in Json Format on to amqp bus</br>
																	/metrics get metric data
																	</body>
																	</html>`))
		})
		APIEndpointURL := fmt.Sprintf("%s", serverConfig.API.APIEndpointURL)
		go func(APIEndpointURL string) {
			log.Printf("APIEndpoint server ready at : %s\n", APIEndpointURL)
			log.Fatal(http.ListenAndServe(APIEndpointURL, nil))
		}(APIEndpointURL)
		time.Sleep(2 * time.Second)
	}
	log.Println("Ready....")

	// start processing  events from QDR
	for {
		select {
		case event := <-amqpEventServer.GetNotifier():
			amqpEventServer.GetHandler().IncTotalMsgProcessed()
			//log.Printf("Event occurred : %#v\n", event)
			indexName, indexType, err := saelastic.GetIndexNameType(event)
			if err != nil {
				log.Printf("Failed to read event %s type in main %s\n", event, err)
				applicationHealth.ElasticSearchState = 0
			} else {
				id, err := elasticClient.Create(indexName, indexType, event)
				applicationHealth.ElasticSearchState = 1
				if err != nil {
					log.Printf("Error creating event %s in elastic search %s\n", event, err)
					applicationHealth.ElasticSearchState = 0
				} // else {
				//update AlertManager
				if serverConfig.AlertManagerEnabled {
					go func() {
						var alert = &alerts.Alerts{}
						var jsonStr = []byte(event)
						generatorURL := fmt.Sprintf("%s/%s/%s/%s", serverConfig.ElasticHostURL, indexName, indexType, id)
						alert.Parse(jsonStr, generatorURL)
						debuge("Debug:Sending alert..%#v\n", alert)
						debuge("Debug:Generator URL %s\n", generatorURL)
						jsonString, err := json.Marshal(*alert)
						if err != nil {
							panic(err)
						}
						var jsonEvent = []byte("[" + string(jsonString) + "]")
						debuge("Debug:Posting to  %#s\n", serverConfig.AlertManagerURL)
						req, err := http.NewRequest("POST", serverConfig.AlertManagerURL, bytes.NewBuffer(jsonEvent))
						req.Header.Set("X-Custom-Header", "smartgateway")
						req.Header.Set("Content-Type", "application/json")
						client := &http.Client{}
						resp, err := client.Do(req)
						if err != nil {
							panic(err)
						}
						defer resp.Body.Close()
						body, _ := ioutil.ReadAll(resp.Body)
						debuge("Debug:response Status:%s\n", resp.Status)
						debuge("Debug:response Headers:%s\n", resp.Header)
						debuge("Debug:response Body:%s\n", string(body))

					}()
					//}
				}
			}
			continue // priority channel
		case status := <-amqpEventServer.GetStatus():
			applicationHealth.QpidRouterState = status
		default:
			//no activity
		}
	}

}
