package cacheutil

import (
	"log"

	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/infrawatch/smart-gateway/internal/pkg/tsdb"
	"github.com/prometheus/client_golang/prometheus"
)

//AddHeartBeat ...
func AddHeartBeat(instance string, value float64, ch chan<- prometheus.Metric) {
	m, err := tsdb.NewHeartBeatMetricByHost(instance, value)
	if err != nil {
		log.Printf("newHeartBeat: %v for %s", err, instance)
	}
	ch <- m
}

//AddMetricsByHostCount ...
func AddMetricsByHostCount(instance string, value float64, ch chan<- prometheus.Metric) {
	m, err := tsdb.AddMetricsByHost(instance, value)
	if err != nil {
		log.Printf("AddMetricsByHost: %v for %s", err, instance)
	}
	ch <- m
}

//FlushPrometheusMetric   generate Prometheus metrics
func (shard *ShardedIncomingDataCache) FlushPrometheusMetric(usetimestamp bool, ch chan<- prometheus.Metric) int {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	minMetricCreated := 0 //..minimum of one metrics created

	for _, dataInterface := range shard.plugin {
		format := saconfig.DataSourceUniversal.String()
		switch metric := dataInterface.(type) {
		case *incoming.CollectdMetric:
			format = metric.DataSource.String()
		case *incoming.CeilometerMetric:
			format = metric.DataSource.String()
		}
		if dataInterface.ISNew() {
			dataInterface.SetNew(false)
			for index := range dataInterface.GetValues() {
				m, err := tsdb.NewPrometheusMetric(usetimestamp, format, dataInterface, index)
				if err != nil {
					log.Printf("newMetric: %v", err)
					continue
				}
				ch <- m
				minMetricCreated++
			}
		} else {
			//clean up if data is not access for max TTL specified
			if shard.Expired() {
				delete(shard.plugin, dataInterface.GetItemKey())
			}
		}

	}
	return minMetricCreated
}

//FlushAllMetrics   Generic Flushing metrics not used.. used only for testing
func (shard *ShardedIncomingDataCache) FlushAllMetrics() {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	for _, dataInterface := range shard.plugin {
		if dataInterface.ISNew() {
			dataInterface.SetNew(false)
			log.Printf("New Metrics %#v\n", dataInterface)
		} else {
			//clean up if data is not access for max TTL specified
			if shard.Expired() {
				delete(shard.plugin, dataInterface.GetItemKey())
				log.Printf("Cleaned up plugin for %s", dataInterface.GetItemKey())
			}
		}
	}
}
