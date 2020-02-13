package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/infrawatch/smart-gateway/internal/pkg/amqp10"
	"github.com/infrawatch/smart-gateway/internal/pkg/cacheutil"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
)

var debugh = func(format string, data ...interface{}) {} // Default no debugging output

type (
	// Timestamp is a helper for (un)marhalling time
	Timestamp time.Time

	// HookMessage is the message we receive from Alertmanager
	HookMessage struct {
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Status            string            `json:"status"`
		Receiver          string            `json:"receiver"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Alerts            []Alert           `json:"alerts"`
	}

	//Alert is a single alert.
	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt,omitempty"`
		EndsAt      string            `json:"EndsAt,omitempty"`
	}

	//Context ...
	Context struct {
		Config      *saconfig.EventConfiguration
		AMQP1Sender *amqp10.AMQPSender
	}

	//Handler ...
	Handler struct {
		*Context
		H func(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
	}
)

//NewContext ...
func NewContext(serverConfig saconfig.EventConfiguration) *Context {
	amqpPublishurl := fmt.Sprintf("amqp://%s", serverConfig.API.AMQP1PublishURL)
	amqpSender := amqp10.NewAMQPSender(amqpPublishurl, false)
	context := &Context{Config: &serverConfig, AMQP1Sender: amqpSender}
	if serverConfig.Debug {
		debugh = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}
	return context
}

//ServeHTTP...
func (ah Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Updated to pass ah.appContext as a parameter to our handler type.
	status, err := ah.H(ah.Context, w, r)
	if err != nil {
		debugh("Debug:HTTP %d: %q", status, err)
		switch status {
		case http.StatusNotFound:
			http.NotFound(w, r)
			// And if we wanted a friendlier error page:
			// err := ah.renderTemplate(w, "http_404.tmpl", nil)
		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(status), status)
		default:
			http.Error(w, http.StatusText(status), status)
		}
	}
}

//AlertHandler  ...
func AlertHandler(a *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	var body HookMessage
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "invalid request body", 400)
		return http.StatusInternalServerError, err
	}

	debugh("API AlertHandler Body%#v\n", body)
	out, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	debugh("Debug:Sending alerts to to AMQP")
	debugh("Debug:Alert on AMQP%#v\n", string(out))
	a.AMQP1Sender.Send(string(out))

	// We can shortcut this: since renderTemplate returns `error`,
	// our ServeHTTP method will return a HTTP 500 instead and won't
	// attempt to write a broken template out with a HTTP 200 status.
	// (see the postscript for how renderTemplate is implemented)
	// If it doesn't return an error, things will go as planned.
	return http.StatusOK, nil
}

//MetricHandler ...metric handlers
type MetricHandler struct {
	applicationHealth *cacheutil.ApplicationHealthCache
	lastPull          *prometheus.Desc
	qpidRouterState   *prometheus.Desc
}

//EventMetricHandler  ....
type EventMetricHandler struct {
	applicationHealth  *cacheutil.ApplicationHealthCache
	lastPull           *prometheus.Desc
	qpidRouterState    *prometheus.Desc
	elasticSearchState *prometheus.Desc
}

//NewAppStateMetricHandler  ...
func NewAppStateMetricHandler(applicationHealth *cacheutil.ApplicationHealthCache) *MetricHandler {
	plabels := prometheus.Labels{}
	plabels["source"] = "Metric Listener"
	return &MetricHandler{
		applicationHealth: applicationHealth,
		lastPull: prometheus.NewDesc("sa_collectd_last_pull_timestamp_seconds",
			"Unix timestamp of the last metrics pull in seconds.",
			nil, plabels,
		),
		qpidRouterState: prometheus.NewDesc("sa_collectd_qpid_router_status",
			"Metric listener router status ",
			nil, plabels,
		),
	}
}

//NewAppStateEventMetricHandler  ...
func NewAppStateEventMetricHandler(applicationHealth *cacheutil.ApplicationHealthCache) *EventMetricHandler {
	plabels := prometheus.Labels{}
	plabels["source"] = "Event Listener"

	return &EventMetricHandler{
		applicationHealth: applicationHealth,
		lastPull: prometheus.NewDesc("sa_collectd_last_pull_timestamp_seconds",
			"Unix timestamp of the last event listener pull in seconds.",
			nil, plabels,
		),
		qpidRouterState: prometheus.NewDesc("sa_collectd_qpid_router_status",
			"Event listener router status ",
			nil, plabels,
		),
		elasticSearchState: prometheus.NewDesc("sa_collectd_elasticsearch_status",
			"Event listener ElasticSearch status ",
			nil, plabels,
		),
	}
}

// Describe implements prometheus.Collector.
func (metricHandler *MetricHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricHandler.lastPull
	ch <- metricHandler.qpidRouterState
}

// Describe implements prometheus.Collector.
func (eventMetricHandler *EventMetricHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- eventMetricHandler.lastPull
	ch <- eventMetricHandler.qpidRouterState
	ch <- eventMetricHandler.elasticSearchState
}

// Collect implements prometheus.Collector.
func (metricHandler *MetricHandler) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(metricHandler.lastPull, prometheus.GaugeValue, float64(time.Now().Unix()))
	ch <- prometheus.MustNewConstMetric(metricHandler.qpidRouterState, prometheus.GaugeValue, float64(metricHandler.applicationHealth.QpidRouterState))
}

// Collect implements prometheus.Collector.
func (eventMetricHandler *EventMetricHandler) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(eventMetricHandler.lastPull, prometheus.GaugeValue, float64(time.Now().Unix()))
	ch <- prometheus.MustNewConstMetric(eventMetricHandler.qpidRouterState, prometheus.GaugeValue, float64(eventMetricHandler.applicationHealth.QpidRouterState))
	ch <- prometheus.MustNewConstMetric(eventMetricHandler.elasticSearchState, prometheus.GaugeValue, float64(eventMetricHandler.applicationHealth.ElasticSearchState))
}
