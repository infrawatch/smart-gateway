package metrics

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/amqp10"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/api"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/cacheutil"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/incoming"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
)

var (
	debugm       = func(format string, data ...interface{}) {} // Default no debugging output
	debugs       = func(count int) {}                          // Default no debugging output
	serverConfig saconfig.MetricConfiguration
)

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache    *cacheutil.IncomingDataCache
	appstate *api.MetricHandler
}

// Describe implements prometheus.Collector.
func (c *cacheHandler) Describe(ch chan<- *prometheus.Desc) {
	c.appstate.Describe(ch)
}

//Collect implements prometheus.Collector.
//need improvement add lock etc etc
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	//lastPull.Set(float64(time.Now().UnixNano()) / 1e9)
	c.appstate.Collect(ch)
	var metricCount int
	//ch <- lastPull
	lock, allHosts := c.cache.GetHosts()
	defer lock.Unlock()
	debugm("Debug:Prometheus is requesting to scrape metrics...")
	for key, plugin := range allHosts {
		//fmt.Fprintln(w, hostname)
		debugm("Debug:Getting metrics for host %s  with total plugin size %d\n", key, plugin.Size())
		metricCount = plugin.FlushPrometheusMetric(serverConfig.UseTimeStamp, ch)
		if metricCount > 0 {
			// add heart if there is atleast one new metrics for the host
			debugm("Debug:Adding heartbeat for host %s.", key)
			cacheutil.AddHeartBeat(key, 1.0, ch)
		} else {
			cacheutil.AddHeartBeat(key, 0.0, ch)
		}
		//add count of metrics
		cacheutil.AddMetricsByHostCount(key, float64(metricCount), ch)
		//this will clean up all zero plugins
		if plugin.Size() == 0 {
			debugm("Debug:Cleaning all empty plugins.")
			debugm("Debug:Deleting host key %s\n", key)
			delete(allHosts, key)
			debugm("Debug:Cleaned up cache for host %s", key)
		}
	}
}

/*************** main routine ***********************/
// metricusage and command-line flags
func metricusage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run cmd/main.go -config sa.metrics.config.json -debug -servicetype metrics
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run cmd/main.go -servicetype metrics -mhost=localhost -mport=8081 -amqp1MetricURL=10.19.110.5:5672/collectd/telemetry
	**************************************************************`)

	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

//StartMetrics ... entry point to metrics
func StartMetrics() {
	// set flags for parsing options
	flag.Usage = metricusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fServiceType := flag.String("servicetype", "metrics", "metric type")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fIncludeStats := flag.Bool("cpustats", false, "Include cpu usage info in http requests (degrades performance)")
	fExporterhost := flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
	fExporterport := flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
	fAMQP1MetricURL := flag.String("amqp1MetricURL", "", "AMQP1.0 metrics listener example 127.0.0.1:5672/collectd/telemetry")
	fCount := flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")
	fPrefetch := flag.Int("prefetch", 0, "AMQP1.0 option: Enable prefetc and set capacity(0 is disabled,>0 enabled with capacity of >0) (OPTIONAL)")

	fUsetimestamp := flag.Bool("usetimestamp", true, "Use source time stamp instead of promethues.(default true,OPTIONAL)")
	fUniqueName := flag.String("uname", "metrics-"+strconv.Itoa(rand.Intn(100)), "Unique name across application")

	flag.Parse()

	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadMetricConfig(*fConfigLocation)
		serverConfig.ServiceType = *fServiceType
		if *fDebug {
			serverConfig.Debug = true
		}
	} else {
		serverConfig = saconfig.MetricConfiguration{
			AMQP1MetricURL: *fAMQP1MetricURL,
			CPUStats:       *fIncludeStats,
			Exporterhost:   *fExporterhost,
			Exporterport:   *fExporterport,
			DataCount:      *fCount, //-1 for ever which is default
			UseTimeStamp:   *fUsetimestamp,
			Debug:          *fDebug,
			Prefetch:       *fPrefetch,
			ServiceType:    *fServiceType,
			UniqueName:     *fUniqueName,
		}

	}
	if serverConfig.Debug {
		debugm = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if len(serverConfig.AMQP1MetricURL) == 0 {
		log.Println("AMQP1 Metrics URL is required")
		metricusage()
		os.Exit(1)
	}

	//Block -starts
	//Set up signal Go routine
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for sig := range signalCh {
			// sig is a ^C, handle it
			log.Printf("caught sig: %+v", sig)
			log.Println("Wait for 2 second to finish processing")
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
	}()

	//Cache sever to process and serve the exporter
	cacheServer := cacheutil.NewCacheServer(cacheutil.MAXTTL, serverConfig.Debug)
	applicationHealth := cacheutil.NewApplicationHealthCache()
	appStateHandler := api.NewAppStateMetricHandler(applicationHealth)
	myHandler := &cacheHandler{cache: cacheServer.GetCache(), appstate: appStateHandler}
	amqpHandler := amqp10.NewAMQPHandler("Metric Consumer")
	prometheus.MustRegister(myHandler, amqpHandler)

	if !serverConfig.CPUStats {
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		prometheus.Unregister(prometheus.NewGoCollector())
	}
	//Set up Metric Exporter
	handler := http.NewServeMux()
	handler.Handle("/metrics", prometheus.Handler())
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                                <head><title>Collectd Exporter</title></head>
                                <body>cacheutil
                                <h1>Collectd Exporter</h1>
                                <p><a href='/metrics'>Metrics</a></p>
                                </body>
                                </html>`))
	})
	// Register pprof handlers
	handler.HandleFunc("/debug/pprof/", pprof.Index)
	handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

	debugm("Debug: Config %#v\n", serverConfig)
	//run exporter fro prometheus to scrape
	go func() {
		metricsURL := fmt.Sprintf("%s:%d", serverConfig.Exporterhost, serverConfig.Exporterport)
		log.Printf("Metric server at : %s\n", metricsURL)
		log.Fatal(http.ListenAndServe(metricsURL, handler))
	}()
	time.Sleep(2 * time.Second)
	log.Println("HTTP server is ready....")

	///Metric Listener
	amqpMetricsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1MetricURL)
	log.Printf("Connecting to AMQP1 : %s\n", amqpMetricsurl)
	amqpMetricServer := amqp10.NewAMQPServer(amqpMetricsurl, serverConfig.Debug, serverConfig.DataCount, serverConfig.Prefetch, amqpHandler, *fUniqueName)
	log.Printf("Listening.....\n")

msgloop:
	for {
		select {
		case data := <-amqpMetricServer.GetNotifier():
			amqpMetricServer.GetHandler().IncTotalMsgProcessed()
			debugm("Debug: Getting incoming data from notifier channel : %#v\n", data)
			incomingType := incoming.NewInComing(incoming.COLLECTD)
			metrics, _ := incomingType.ParseInputJSON(data)
			for _, m := range metrics {
				amqpMetricServer.UpdateMinCollectInterval(m.GetInterval())
				cacheServer.Put(m)
			}
			debugs(len(metrics))
			continue // priority channel
		case status := <-amqpMetricServer.GetStatus():
			applicationHealth.QpidRouterState = status
		case <-amqpMetricServer.GetDoneChan():
			break msgloop
		}
	}
	//TODO: to close cache server on keyboard interrupt
}
