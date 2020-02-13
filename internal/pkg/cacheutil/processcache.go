package cacheutil

import (
	"log"

	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
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

	for _, IncomingDataInterface := range shard.plugin {
		if collectd, ok := IncomingDataInterface.(*incoming.CollectdMetric); ok {
			if collectd.ISNew() {
				collectd.SetNew(false)
				for index := range collectd.Values {
					m, err := tsdb.NewCollectdMetric(usetimestamp, *collectd, index)
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
					delete(shard.plugin, collectd.GetItemKey())
					//log.Printf("Cleaned up plugin for %s", collectd.GetKey())
				}
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
		if collectd, ok := dataInterface.(*incoming.CollectdMetric); ok {
			if collectd.ISNew() {
				collectd.SetNew(false)
				log.Printf("New Metrics %#v\n", collectd)
			} else {
				//clean up if data is not access for max TTL specified
				if shard.Expired() {
					delete(shard.plugin, collectd.GetItemKey())
					log.Printf("Cleaned up plugin for %s", collectd.GetItemKey())
				}
			}
		}
	}
}
