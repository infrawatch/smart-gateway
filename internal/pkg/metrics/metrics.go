package metrics

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/infrawatch/smart-gateway/internal/pkg/amqp10"
	"github.com/infrawatch/smart-gateway/internal/pkg/api"
	"github.com/infrawatch/smart-gateway/internal/pkg/cacheutil"
	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//MetricHandlerHTML contains HTML for default endpoint
const MetricHandlerHTML = `
<html>
		<head><title>Collectd Exporter</title></head>
		<body>
			<h1>Collectd Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
		</body>
</html>
`

var (
	debugm = func(format string, data ...interface{}) {} // Default no debugging output
	debugs = func(count int) {}                          // Default no debugging output
)

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	useTimestamp bool
	cache        *cacheutil.IncomingDataCache
	appstate     *api.MetricHandler
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
		debugm("Debug:Getting metrics for host %s  with total plugin size %d\n", key, plugin.Size())
		metricCount = plugin.FlushPrometheusMetric(c.useTimestamp, ch)
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
	********************* config *********************
	$go run cmd/main.go -config smartgateway_config.json -servicetype metrics
	**************************************************`)

	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

//StartMetrics ... entry point to metrics
func StartMetrics() {
	var wg sync.WaitGroup
	finish := make(chan bool)

	amqp10.SpawnSignalHandler(finish, os.Interrupt)

	// set flags for parsing options
	flag.Usage = metricusage
	fServiceType := flag.String("servicetype", "metrics", "Metric type")
	fConfigLocation := flag.String("config", "", "Path to configuration file.")
	fUniqueName := flag.String("uname", "metrics-"+strconv.Itoa(rand.Intn(100)), "Unique name across application")
	flag.Parse()

	var serverConfig *saconfig.MetricConfiguration
	if len(*fConfigLocation) > 0 { //load configuration
		conf, err := saconfig.LoadConfiguration(*fConfigLocation, "metric")
		if err != nil {
			log.Fatal("Config Parse Error: ", err)
		}
		serverConfig = conf.(*saconfig.MetricConfiguration)
		serverConfig.ServiceType = *fServiceType
	} else {
		metricusage()
		os.Exit(1)
	}
	if serverConfig.Debug {
		debugm = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if len(serverConfig.AMQP1MetricURL) == 0 && len(serverConfig.AMQP1Connections) == 0 {
		log.Println("Configuration option 'AMQP1MetricURL' or 'AMQP1Connections' is required")
		metricusage()
		os.Exit(1)
	}

	if len(serverConfig.AMQP1MetricURL) > 0 {
		serverConfig.AMQP1Connections = []saconfig.AMQPConnection{
			saconfig.AMQPConnection{
				URL:          serverConfig.AMQP1MetricURL,
				DataSourceID: saconfig.DataSourceCollectd,
				DataSource:   "collectd",
			},
		}
	}

	for _, conn := range serverConfig.AMQP1Connections {
		log.Printf("AMQP1.0 %s listen address configured at %s\n", conn.DataSource, conn.URL)
	}

	applicationHealth := cacheutil.NewApplicationHealthCache()
	metricHandler := api.NewAppStateMetricHandler(applicationHealth)
	amqpHandler := amqp10.NewAMQPHandler("Metric Consumer")
	//Cache sever to process and serve the exporter
	cacheServer := cacheutil.NewCacheServer(cacheutil.MAXTTL, serverConfig.Debug)
	cacheHandler := &cacheHandler{useTimestamp: serverConfig.UseTimeStamp, cache: cacheServer.GetCache(), appstate: metricHandler}
	prometheus.MustRegister(cacheHandler, amqpHandler)

	if !serverConfig.CPUStats {
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		prometheus.Unregister(prometheus.NewGoCollector())
	}
	//Set up Metric Exporter
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(MetricHandlerHTML))
	})
	// Register pprof handlers
	handler.HandleFunc("/debug/pprof/", pprof.Index)
	handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

	debugm("Debug: Config %#v\n", serverConfig)
	//run exporter for prometheus to scrape
	go func() {
		metricsURL := fmt.Sprintf("%s:%d", serverConfig.Exporterhost, serverConfig.Exporterport)
		log.Printf("Metric server at : %s\n", metricsURL)
		log.Fatal(http.ListenAndServe(metricsURL, handler))
	}()
	time.Sleep(2 * time.Second)
	log.Println("HTTP server is ready....")

	// AMQP connection(s)
	processingCases, qpidStatusCases, amqpServers := amqp10.CreateMessageLoopComponents(serverConfig, finish, amqpHandler, *fUniqueName)
	amqp10.SpawnQpidStatusReporter(&wg, applicationHealth, qpidStatusCases)

	// spawn metric processor
	wg.Add(1)
	go func() {
		defer wg.Done()
		finishCase := len(processingCases) - 1
	processingLoop:
		for {
			switch index, msg, _ := reflect.Select(processingCases); index {
			case finishCase:
				break processingLoop
			default:
				debugm("Debug: Getting incoming data from notifier channel : %#v\n", msg)
				metric := incoming.NewFromDataSource(amqpServers[index].DataSource)
				amqpServers[index].Server.GetHandler().IncTotalMsgProcessed()
				metrics, _ := metric.ParseInputJSON(msg.String())
				for _, m := range metrics {
					amqpServers[index].Server.UpdateMinCollectInterval(m.GetInterval())
					cacheServer.Put(m)
				}
				debugs(len(metrics))
			}
		}
		log.Println("Closing event processor.")
	}()

	// do not end until all loop goroutines ends
	wg.Wait()
	log.Println("Exiting")
}
