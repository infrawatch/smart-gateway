package saconfig

import (
	"testing"
)

func TestMetricConfig(t *testing.T) {
	var configuration MetricConfiguration
	configuration = LoadMetricConfig("config.sa.metrics.sample.json")
	if len(configuration.AMQP1MetricURL) == 0 {
		t.Error("Empty configuration generated")
	}
	if configuration.Prefetch != 100 {
		t.Error("Error loading prefetch")
	}

}

func TestEventConfig(t *testing.T) {
	var configuration EventConfiguration
	configuration = LoadEventConfig("config.sa.events.sample.json")
	if len(configuration.AMQP1EventURL) == 0 {
		t.Error("Empty configuration generated")
	}
	if len(configuration.AlertManagerURL) == 0 {
		t.Error("Empty configuration generated")
	}
	if configuration.Prefetch != 100 {
		t.Error("Error loading prefetch")
	}
}
