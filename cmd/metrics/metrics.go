package main

import (
	"os/signal"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/api"
	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"

	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
	"time"
)

var debugm = func(format string, data ...interface{}) {} // Default no debugging output

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache    *cacheutil.IncomingDataCache
	appstate *apihandler.MetricHandler
}

// Describe implements prometheus.Collector.
func (c *cacheHandler) Describe(ch chan<- *prometheus.Desc) {
	c.appstate.Describe(ch)
}

// Collect implements prometheus.Collector.
//need improvement add lock etc etc
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	//lastPull.Set(float64(time.Now().UnixNano()) / 1e9)
	c.appstate.Collect(ch)
	//ch <- lastPull
	allHosts := c.cache.GetHosts()
	debugm("Debug:Prometheus is requesting to scrape metrics...")
	for key, plugin := range allHosts {
		//fmt.Fprintln(w, hostname)
		debugm("Debug:Getting metrics for host %s  with total plugin size %d\n", key, plugin.Size())
		if plugin.FlushPrometheusMetric(ch) == true {
			// add heart if there is atleast one new metrics for the host
			debugm("Debug:Adding heartbeat for host %s.", key)
			cacheutil.AddHeartBeat(key, 1.0, ch)
		} else {
			cacheutil.AddHeartBeat(key, 0.0, ch)
		}

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
	$go run metrics/main.go -config sa.metrics.config.json -debug
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run metrics/main.go -mhost=localhost -mport=8081 -amqp1MetricURL=10.19.110.5:5672/collectd/telemetry
	**************************************************************

	For running Sample data wihout AMQP use following option
	********************* Sample Data *********************
	$go run metrics/main.go -mhost=localhost -mport=8081 -usesample=true -h=10 -p=100 -t=-1 -debug
	*************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func main() {
	// set flags for parsing options
	flag.Usage = metricusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fIncludeStats := flag.Bool("cpustats", false, "Include cpu usage info in http requests (degrades performance)")
	fExporterhost := flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
	fExporterport := flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
	fAMQP1MetricURL := flag.String("amqp1MetricURL", "", "AMQP1.0 metrics listener example 127.0.0.1:5672/collectd/telemetry")
	fCount := flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")

	fSampledata := flag.Bool("usesample", false, "Use sample data instead of amqp.This will not fetch any data from amqp (OPTIONAL)")
	fHosts := flag.Int("h", 1, "No of hosts : Sample hosts required (default 1).")
	fPlugins := flag.Int("p", 100, "No of plugins: Sample plugins per host(default 100).")
	fIterations := flag.Int("t", 1, "No of times to run sample data (default 1) -1 for ever.")

	flag.Parse()
	var serverConfig saconfig.MetricConfiguration
	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadMetricConfig(*fConfigLocation)
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
			UseSample:      *fSampledata,
			Debug:          *fDebug,
			Sample: saconfig.SampleDataConfig{
				HostCount:   *fHosts,   //no of host to simulate
				PluginCount: *fPlugins, //No of plugin count per hosts
				DataCount:   *fIterations,
			},
		}

	}
	if serverConfig.Debug {
		debugm = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if serverConfig.UseSample == false && (len(serverConfig.AMQP1MetricURL) == 0) {
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
	type MetricHandler struct {
		applicationHealth *cacheutil.ApplicationHealthCache
		lastPull          *prometheus.Desc
		qpidRouterState   *prometheus.Desc
	}

	applicationHealth := cacheutil.NewApplicationHealthCache()
	appStateHandler := apihandler.NewAppStateMetricHandler(applicationHealth)
	myHandler := &cacheHandler{cache: cacheServer.GetCache(), appstate: appStateHandler}
	prometheus.MustRegister(myHandler)

	if serverConfig.CPUStats == false {
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
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

	//if running just samples
	if serverConfig.UseSample {
		log.Println("Using sample data")
		if serverConfig.Sample.DataCount == -1 {
			serverConfig.Sample.DataCount = 9999999
		}
		var hostwaitgroup sync.WaitGroup
		fmt.Printf("Test data  will run for %d times ", serverConfig.Sample.DataCount)
		for times := 1; times <= serverConfig.Sample.DataCount; times++ {
			hostwaitgroup.Add(serverConfig.Sample.HostCount)
			for hosts := 0; hosts < serverConfig.Sample.HostCount; hosts++ {
				go func(host_id int) {
					defer hostwaitgroup.Done()
					hostname := fmt.Sprintf("%s_%d", "redhat.boston.nfv", host_id)
					incomingType := incoming.NewInComing(incoming.COLLECTD)
					debugm("Hostname %s IncomingType %#v", hostname, incomingType)
					go cacheServer.GenrateSampleData(hostname, serverConfig.Sample.PluginCount, incomingType)
				}(hosts)

			}
			hostwaitgroup.Wait()
			time.Sleep(time.Second * 1)
		}

	} else {
		//aqp listener if sample is requested then amqp will not be used but random sample data will be used
		var amqpMetricServer *amqp10.AMQPServer
		///Metric Listener
		amqpMetricsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1MetricURL)
		log.Printf("Connecting to AMQP1 : %s\n", amqpMetricsurl)
		amqpMetricServer = amqp10.NewAMQPServer(amqpMetricsurl, serverConfig.Debug, serverConfig.DataCount)
		log.Printf("Listening.....\n")

		for {
			select {
			case data := <-amqpMetricServer.GetNotifier():
				debugm("Debug: Getting incoming data from notifier channel : %#v\n", data)
				incomingType := incoming.NewInComing(incoming.COLLECTD)
				incomingType.ParseInputJSON(data)
				cacheServer.Put(incomingType)
				continue // priority channel
			case status := <-amqpMetricServer.GetStatus():
				applicationHealth.QpidRouterState = status
			default:
				//no activity
			}
		}
	}
	//TO DO: to close cache server on keyboard interrupt

}
