package tsdb

import (
	"fmt"
	"regexp"
	"time"

	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	isoTimeLayout = "2006-01-02 15:04:05.000000"
	RFC3339Python = "2006-01-02T15:04:05.000000"
)

//TSDB  interface
type TSDB interface {
	//prometheus specific reflect
	GetLabels() map[string]string
	GetMetricName(index int) string
	GetMetricDesc(index int) string
}

var (
	metricNameRe = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

//NewHeartBeatMetricByHost ...
func NewHeartBeatMetricByHost(instance string, value float64) (prometheus.Metric, error) {
	valueType := prometheus.GaugeValue
	metricName := "collectd_last_metric_for_host_status"
	help := "Status of metrics for host currently active."

	plabels := prometheus.Labels{}
	plabels["instance"] = instance
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)
	return prometheus.NewConstMetric(desc, valueType, value)
}

//AddMetricsByHost ...
func AddMetricsByHost(instance string, value float64) (prometheus.Metric, error) {
	valueType := prometheus.GaugeValue
	metricName := "collectd_metric_per_host"
	help := "No of metrics for host currently read."

	plabels := prometheus.Labels{}
	plabels["instance"] = instance
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)
	return prometheus.NewConstMetric(desc, valueType, value)
}

//NewPrometheusMetric converts one data source of a value list to a Prometheus metric.
func NewPrometheusMetric(usetimestamp bool, format string, metric incoming.MetricDataFormat, index int) (prometheus.Metric, error) {
	var (
		timestamp        time.Time
		valueType        prometheus.ValueType
		metricName, help string
		labels           map[string]string
		value            float64
	)

	if format == saconfig.DataSourceCollectd.String() {
		collectd := metric.(*incoming.CollectdMetric)
		switch collectd.Dstypes[index] {
		case "gauge":
			valueType = prometheus.GaugeValue
		case "derive", "counter":
			valueType = prometheus.CounterValue
		default:
			return nil, fmt.Errorf("unknown name of value type: %s", collectd.Dstypes[index])
		}
		timestamp = collectd.Time.Time()
		help = collectd.GetMetricDesc(index)
		metricName = metricNameRe.ReplaceAllString(collectd.GetMetricName(index), "_")
		labels = collectd.GetLabels()
		value = collectd.Values[index]
	} else if format == saconfig.DataSourceCeilometer.String() {
		ceilometer := metric.(*incoming.CeilometerMetric)
		if ctype, ok := ceilometer.Payload["counter_type"]; ok {
			if counterType, ok := ctype.(string); ok {
				switch counterType {
				case "gauge":
					valueType = prometheus.GaugeValue
				default:
					valueType = prometheus.CounterValue
				}
			} else {
				return nil, fmt.Errorf("invalid counter_type in metric payload: %s", ceilometer.Payload)
			}
		} else {
			return nil, fmt.Errorf("did not find counter_type in metric payload: %s", ceilometer.Payload)
		}
		if ts, ok := ceilometer.Payload["timestamp"]; ok {
 			for _, layout := range []string{time.RFC3339, time.RFC3339Nano, time.ANSIC, RFC3339Python, isoTimeLayout} {
				if stamp, err := time.Parse(layout, ts.(string)); err == nil {
					timestamp = stamp
					break
			}
			if timestamp.IsZero() {
				return nil, fmt.Errorf("invalid timestamp in metric payload: %s", ceilometer.Payload)
			}
		} else {
			return nil, fmt.Errorf("did not find timestamp in metric payload: %s", ceilometer.Payload)
		}
		help = ceilometer.GetMetricDesc(index)
		metricName = metricNameRe.ReplaceAllString(ceilometer.GetMetricName(index), "_")
		labels = ceilometer.GetLabels()
		value = ceilometer.Values[index]
	}

	plabels := prometheus.Labels{}
	for key, value := range labels {
		plabels[key] = value
	}
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)
	if usetimestamp {
		return prometheus.NewMetricWithTimestamp(
			timestamp,
			prometheus.MustNewConstMetric(desc, valueType, value),
		), nil
	}
	return prometheus.NewConstMetric(desc, valueType, value)
}
