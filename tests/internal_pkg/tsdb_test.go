package tests

import (
	"strings"
	"testing"

	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/tsdb"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

/*----------------------------- helper functions -----------------------------*/

//GenerateSampleCacheData  ....
func GenerateCollectdMetric(hostname string, pluginname string, useTimestamp bool, index int) (*incoming.CollectdMetric, prometheus.Metric, dto.Metric) {
	sample := GenerateSampleCollectdData(hostname, pluginname)
	collectdMetric, _ := tsdb.NewCollectdMetric(useTimestamp, *sample, index)
	metric := dto.Metric{}
	collectdMetric.Write(&metric)
	return sample, collectdMetric, metric
}

/*----------------------------------------------------------------------------*/

func TestTimestamp(t *testing.T) {
	sample, _, metric := GenerateCollectdMetric("hostname", "pluginname", true, 0)
	assert.Equal(t, sample.Time.Time().UnixNano()/1000000, metric.GetTimestampMs())

	sample, _, metric = GenerateCollectdMetric("hostname", "pluginname", false, 0)
	assert.NotEqual(t, sample.Time.Time().UnixNano()/1000000, metric.GetTimestampMs())
}

func TestCollectdMetric(t *testing.T) {
	t.Run("Test prometeus metric values", func(t *testing.T) {
		sample, collectdMetric, metric := GenerateCollectdMetric("hostname", "pluginname", true, 0)
		assert.True(t, strings.HasPrefix(collectdMetric.Desc().String(), "Desc{fqName: \"collectd_pluginname_collectd_value1\""))
		assert.Equal(t, sample.Values[0], metric.GetGauge().GetValue())
		assert.Equal(t, 0.0, metric.GetCounter().GetValue())

		sample, collectdMetric, metric = GenerateCollectdMetric("hostname", "pluginname", true, 1)
		assert.True(t, strings.HasPrefix(collectdMetric.Desc().String(), "Desc{fqName: \"collectd_pluginname_collectd_value2_total\""))
		assert.Equal(t, sample.Values[1], metric.GetCounter().GetValue())
		assert.Equal(t, 0.0, metric.GetGauge().GetValue())
	})

	t.Run("Test heart beat metric", func(t *testing.T) {
		collectdMetric, _ := tsdb.NewHeartBeatMetricByHost("test_heartbeat", 66.6)
		assert.True(t, strings.HasPrefix(collectdMetric.Desc().String(), "Desc{fqName: \"collectd_last_metric_for_host_status\""))
		metric := dto.Metric{}
		collectdMetric.Write(&metric)
		assert.Equal(t, 66.6, metric.GetGauge().GetValue())
		assert.Equal(t, "test_heartbeat", metric.GetLabel()[0].GetValue())
	})

	t.Run("Test metric by host", func(t *testing.T) {
		collectdMetric, _ := tsdb.AddMetricsByHost("test_host", 666.0)
		assert.True(t, strings.HasPrefix(collectdMetric.Desc().String(), "Desc{fqName: \"collectd_metric_per_host\""))
		metric := dto.Metric{}
		collectdMetric.Write(&metric)
		assert.Equal(t, 666.0, metric.GetGauge().GetValue())
		assert.Equal(t, "test_host", metric.GetLabel()[0].GetValue())
	})
}
