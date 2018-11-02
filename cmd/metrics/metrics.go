package main

import (
	"math/rand"
	"os/signal"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/amqp"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/api"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/cacheutil"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/config"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/incoming"

	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
	"time"
)

var (
	cacheServer      cacheutil.CacheServer
	amqpMetricServer *amqp10.AMQPServer
	debugm           = func(format string, data ...interface{}) {} // Default no debugging output
	debugs           = func(count int) {}                          // Default no debugging output
	serverConfig     saconfig.MetricConfiguration
	amqpHandler      *amqp10.AMQPHandler
	myHandler        *cacheHandler
	hostwaitgroup    sync.WaitGroup
)

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
	var metricCount int
	//ch <- lastPull
	allHosts := c.cache.GetHosts()
	debugm("Debug:Prometheus is requesting to scrape metrics...")
	for key, plugin := range allHosts {
		//fmt.Fprintln(w, hostname)
		debugm("Debug:Getting metrics for host %s  with total plugin size %d\n", key, plugin.Size())
		metricCount = plugin.FlushPrometheusMetric(ch)
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

func getLoopStater(q chan string, everyCount int) func(count int) {
	lastCounted := time.Now()
	showCount := everyCount
	var lenSum float64
	var countSent int64

	return func(count int) {
		countSent = countSent + int64(count)
		lenSum = lenSum + float64(len(q))
		if showCount != -1 && countSent%int64(showCount+1) == 0 {
			lastCounted = time.Now()
		}
		if showCount != -1 && countSent%int64(showCount) == 0 {
			deltaTime := time.Now().Sub(lastCounted)
			msPerMetric := float64(deltaTime/time.Microsecond) / float64(showCount)
			log.Printf("Rcv: %d Rcv'd (%v %.4v us), queue depth %v\n", countSent, deltaTime, msPerMetric, lenSum/float64(countSent))
		}
	}
}
func main() {
	// set flags for parsing options
	flag.Usage = metricusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fTestServer := flag.Bool("testclient", false, "Enable Test Receiver for use with AMQP test client")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fIncludeStats := flag.Bool("cpustats", false, "Include cpu usage info in http requests (degrades performance)")
	fExporterhost := flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
	fExporterport := flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
	fAMQP1MetricURL := flag.String("amqp1MetricURL", "", "AMQP1.0 metrics listener example 127.0.0.1:5672/collectd/telemetry")
	fCount := flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")
	fPrefetch := flag.Int("prefetch", 0, "AMQP1.0 option: Enable prefetc and set capacity(0 is disabled,>0 enabled with capacity of >0) (OPTIONAL)")

	fSampledata := flag.Bool("usesample", false, "Use sample data instead of amqp.This will not fetch any data from amqp (OPTIONAL)")
	fHosts := flag.Int("h", 1, "No of hosts : Sample hosts required (default 1).")
	fPlugins := flag.Int("p", 100, "No of plugins: Sample plugins per host(default 100).")
	fIterations := flag.Int("t", 1, "No of times to run sample data (default 1) -1 for ever.")
	fUniqueName := flag.String("uname", "metrics-"+strconv.Itoa(rand.Intn(100)), "Unique name across application")

	flag.Parse()

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
			TestServer:     *fTestServer,
			Prefetch:       *fPrefetch,
			Sample: saconfig.SampleDataConfig{
				HostCount:   *fHosts,   //no of host to simulate
				PluginCount: *fPlugins, //No of plugin count per hosts
				DataCount:   *fIterations,
			},
			UniqueName: *fUniqueName,
		}

	}
	if serverConfig.Debug {
		debugm = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if serverConfig.TestServer == false &&
		serverConfig.UseSample == false && (len(serverConfig.AMQP1MetricURL) == 0) {
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
	appStateHandler := apihandler.NewAppStateMetricHandler(applicationHealth)
	myHandler := &cacheHandler{cache: cacheServer.GetCache(), appstate: appStateHandler}
	amqpHandler := amqp10.NewAMQPHandler("Metric Consumer")
	prometheus.MustRegister(myHandler, amqpHandler)

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

		// done channel is used during serverTest
		done := make(chan bool)

		///Metric Listener
		amqpMetricsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1MetricURL)
		log.Printf("Connecting to AMQP1 : %s\n", amqpMetricsurl)
		amqpMetricServer = amqp10.NewAMQPServer(amqpMetricsurl, serverConfig.Debug, serverConfig.DataCount, serverConfig.Prefetch, amqpHandler, done, *fTestServer, *fUniqueName)
		log.Printf("Listening.....\n")

		if serverConfig.TestServer == true {
			debugs = getLoopStater(amqpMetricServer.GetNotifier(), 10000)
		}

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
			case <-done:
				break msgloop
			}
		}
	}
	//TO DO: to close cache server on keyboard interrupt

}
