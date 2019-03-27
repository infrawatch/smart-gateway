package tsdb

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/incoming"
)

var (
	metricNameRe = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

//NewHeartBeatMetricByHost ...
func NewHeartBeatMetricByHost(instance string, value float64) (prometheus.Metric, error) {
	var valueType prometheus.ValueType
	valueType = prometheus.GaugeValue
	metricName := "sa_collectd_last_metric_for_host_status"
	help := "Status of metrics for host currently active."

	plabels := prometheus.Labels{}
	plabels["instance"] = instance
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)
	return prometheus.NewConstMetric(desc, valueType, value)
}

//AddMetricsByHost ...
func AddMetricsByHost(instance string, value float64) (prometheus.Metric, error) {
	var valueType prometheus.ValueType
	valueType = prometheus.GaugeValue
	metricName := "sa_collectd_metric_per_host"
	help := "No of metrics for host currently read."

	plabels := prometheus.Labels{}
	plabels["instance"] = instance
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)
	return prometheus.NewConstMetric(desc, valueType, value)
}

//NewCollectdMetric converts one data source of a value list to a Prometheus metric.
func NewCollectdMetric(usetimestamp bool, collectd incoming.Collectd, index int) (prometheus.Metric, error) {
	var value float64
	var valueType prometheus.ValueType

	switch collectd.Dstypes[index] {
	case "gauge":
		value = float64(collectd.Values[index])
		valueType = prometheus.GaugeValue
	case "derive", "counter":
		value = float64(collectd.Values[index])
		valueType = prometheus.CounterValue
	default:
		return nil, fmt.Errorf("unknowdsnamen value type: %s", collectd.Dstypes[index])
	}
	labels := collectd.GetLabels()
	plabels := prometheus.Labels{}
	for key, value := range labels {
		plabels[key] = value
	}

	help := fmt.Sprintf("Service Assurance exporter: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		collectd.Plugin, collectd.Type, collectd.Dstypes[index], collectd.DSName(index))
	metricName := metricNameRe.ReplaceAllString(collectd.GetMetricName(index), "_")
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)

	if usetimestamp == true {
		return prometheus.NewMetricWithTimestamp(
			collectd.Time.Time(),
			prometheus.MustNewConstMetric(desc, valueType, value),
		), nil
	}
	return prometheus.NewConstMetric(desc, valueType, value)

}
