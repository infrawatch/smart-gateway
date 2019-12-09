package events

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/amqp10"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/api"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/cacheutil"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/events/incoming"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic"
)

const (
	EVENTSINDEXTYPE = "event"
	APIHOME         = `
<html>
	<head>
		<title>Smart Gateway Event API</title>
	</head>
	<body>
		<h1>API</h1>
		<ul>
			<li>/alerts POST alerts in JSON format on to AMQP message bus</li>
			<li>/metrics GET metric data</li>
		</ul>
	</body>
</html>
`
)

type AMQPServerItem struct {
	Server     *amqp10.AMQPServer
	DataSource saconfig.DataSource
}

/*************** main routine ***********************/
// eventusage and command-line flags
func eventusage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run cmd/main.go -config sa.events.config.json -debug -servicetype events
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run cmd/main.go -servicetype events -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200
	**************************************************************
	For running with AMQP ,Prometheus,API and AlertManager use following option
	********************* Production *********************
	$go run cmd/main.go -servicetype events -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200 -alertmanager=http://localhost:9090/v1/api/alert -apiurl=localhost:8082 -amqppublishurl=127.0.0.1:5672/collectd/alert
	**************************************************************`)
	fmt.Fprintln(os.Stderr, `Required command line argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

var (
	debuge          = func(format string, data ...interface{}) {} // Default no debugging output
	amqpEventServer *amqp10.AMQPServer
	serverConfig    saconfig.EventConfiguration
)

//spawnSignalHandler spawns goroutine which will wait for interruption signal(s)
// and end smart gateway in case any of the signal is received
func spawnSignalHandler(watchedSignals ...os.Signal) {
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, watchedSignals...)
	go func() {
		for sig := range interruptChannel {
			log.Printf("Stopping execution on caught signal: %+v\n", sig)
			//TO-DO(mmagr): Don't wait based on time, but implement channels to report finished state
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
	}()
}

//StartEvents is the entry point for running smart-gateway in events mode
func StartEvents() {
	spawnSignalHandler(os.Interrupt)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// set flags for parsing options
	flag.Usage = eventusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fServiceType := flag.String("servicetype", "event", "event type")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fAMQP1EventURL := flag.String("amqp1EventURL", "", "AMQP1.0 collectd events listener example 127.0.0.1:5672/collectd/notify")
	fElasticHostURL := flag.String("eshost", "", "ElasticSearch host http://localhost:9200")
	fAlertManagerURL := flag.String("alertmanager", "", "(Optional)AlertManager endpoint http://localhost:9090/v1/api/alert")
	fAPIEndpointURL := flag.String("apiurl", "", "(Optional)API endpoint localhost:8082")
	fAMQP1PublishURL := flag.String("amqppublishurl", "", "(Optional) AMQP1.0 event publish address 127.0.0.1:5672/collectd/alert")
	fResetIndex := flag.Bool("resetIndex", false, "Optional Clean all index before on start (default false)")
	fPrefetch := flag.Int("prefetch", 0, "AMQP1.0 option: Enable prefetc and set capacity(0 is disabled,>0 enabled with capacity of >0) (OPTIONAL)")
	fUniqueName := flag.String("uname", "metrics-"+strconv.Itoa(rand.Intn(100)), "Unique name across application")
	flag.Parse()

	//load configuration from given config file or from cmdline parameters
	if len(*fConfigLocation) > 0 {
		conf, err := saconfig.LoadConfiguration(*fConfigLocation, "event")
		if err != nil {
			log.Fatal("Config Parse Error: ", err)
		}
		serverConfig = conf.(saconfig.EventConfiguration)
		serverConfig.ServiceType = *fServiceType
		if *fDebug {
			serverConfig.Debug = true
		}
	} else {
		serverConfig = saconfig.EventConfiguration{
			AMQP1EventURL:   *fAMQP1EventURL,
			ElasticHostURL:  *fElasticHostURL,
			AlertManagerURL: *fAlertManagerURL,
			Prefetch:        *fPrefetch,
			ServiceType:     *fServiceType,
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

	if len(serverConfig.AMQP1EventURL) == 0 && len(serverConfig.AMQP1Connections) == 0 {
		log.Println("Configuration option 'AMQP1EventURL' or 'AMQP1Connections' is required")
		eventusage()
		os.Exit(1)
	}

	if len(serverConfig.ElasticHostURL) == 0 {
		log.Println("Configuration option 'ElasticHostURL' is required")
		eventusage()
		os.Exit(1)
	}

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

	if len(serverConfig.AMQP1EventURL) > 0 {
		//TO-DO(mmagr): Remove this in next major release
		serverConfig.AMQP1Connections = []saconfig.AMQPConnection{
			saconfig.AMQPConnection{
				Url:          serverConfig.AMQP1EventURL,
				DataSourceId: saconfig.DATA_SOURCE_COLLECTD,
				DataSource:   "collectd",
			},
		}
	}

	applicationHealth := cacheutil.NewApplicationHealthCache()
	metricHandler := api.NewAppStateEventMetricHandler(applicationHealth)
	amqpHandler := amqp10.NewAMQPHandler("Event Consumer")

	// Elastic connection
	elasticClient, err := saelastic.CreateClient(serverConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Connected to ElasticSearch at '%s'.\n", serverConfig.ElasticHostURL)
	applicationHealth.ElasticSearchState = 1

	// API spawn
	if serverConfig.APIEnabled {
		prometheus.MustRegister(metricHandler, amqpHandler)
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		prometheus.Unregister(prometheus.NewGoCollector())

		context := api.NewContext(serverConfig)
		http.Handle("/alert", api.Handler{Context: context, H: api.AlertHandler})
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(APIHOME))
		})

		go func() {
			APIEndpointURL := serverConfig.API.APIEndpointURL
			log.Printf("API server is listening at '%s'\n", APIEndpointURL)
			log.Fatal(http.ListenAndServe(APIEndpointURL, nil))
		}()
	}

	// AMQP connection(s)
	processingCases := make([]reflect.SelectCase, 0, len(serverConfig.AMQP1Connections))
	qpidStatusCases := make([]reflect.SelectCase, 0, len(serverConfig.AMQP1Connections))
	amqpServers := make([]AMQPServerItem, 0, len(serverConfig.AMQP1Connections))
	for _, conn := range serverConfig.AMQP1Connections {
		amqpServer := amqp10.NewAMQPServer(conn.Url, serverConfig.Debug, -1, serverConfig.Prefetch, amqpHandler, *fUniqueName)
		log.Printf("Listening for AMQP messages at '%s'.\n", conn.Url)
		//create select case for this listener
		processingCases = append(processingCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(amqpServer.GetNotifier()),
		})
		qpidStatusCases = append(qpidStatusCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(amqpServer.GetStatus()),
		})
		amqpServers = append(amqpServers, AMQPServerItem{amqpServer, conn.DataSourceId})
	}
	// spawn QPID status reporter
	go func() {
		for {
			_, status, _ := reflect.Select(qpidStatusCases)
			// Note: status here is always very low integer, so we don't need to be afraid of int64>int conversion
			applicationHealth.QpidRouterState = int(status.Int())
		}
	}()
	// spawn event processor
	go func() {
		for {
			index, msg, _ := reflect.Select(processingCases)
			var event incoming.EventDataFormat
			switch amqpServers[index].DataSource.String() {
			case "collectd":
				event = &incoming.CollectdEvent{}
			case "ceilometer":
				// noop for now, gonna panic if configured
				//event = incoming.CeilometerEvent{}
				log.Printf("Received Ceilometer event:\n%s\n", msg)
			case "generic":
				// noop for now, gonna panic if configured
				//event = incoming.GenericEvent{}
				log.Printf("Received generic event:\n%s\n", msg)
			}
			amqpServers[index].Server.GetHandler().IncTotalMsgProcessed()
			err := event.ParseEvent(msg.String())
			if err != nil {
				log.Printf("Failed to parse received event:\n- error: %s\n- event: %s\n", err, event)
			}

			record, err := elasticClient.Create(event.GetIndexName(), EVENTSINDEXTYPE, event.GetSanitized())
			if err != nil {
				applicationHealth.ElasticSearchState = 0
				log.Printf("Failed to save event to Elasticsearch DB:\n- error: %s\n- event: %s\n", err, event)
			} else {
				applicationHealth.ElasticSearchState = 1
			}
			if serverConfig.AlertManagerEnabled {
				go func() {
					generatorUrl := fmt.Sprintf("%s/%s/%s/%s", serverConfig.ElasticHostURL, event.GetIndexName(), EVENTSINDEXTYPE, record)
					alert, err := event.GeneratePrometheusAlertBody(generatorUrl)
					debuge("Debug: Generated alert:\n%s\n", alert)
					var byteAlertBody = []byte(fmt.Sprintf("[%s]", alert))
					req, _ := http.NewRequest("POST", serverConfig.AlertManagerURL, bytes.NewBuffer(byteAlertBody))
					req.Header.Set("X-Custom-Header", "smartgateway")
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						log.Printf("Failed to report alert to AlertManager:\n- error: %s\n- alert: %s\n", err, alert)
						body, _ := ioutil.ReadAll(resp.Body)
						defer resp.Body.Close()
						debuge("Debug:response Status:%s\n", resp.Status)
						debuge("Debug:response Headers:%s\n", resp.Header)
						debuge("Debug:response Body:%s\n", string(body))
					}
				}()
			}
		}
	}()
}
