package tsdb

import (
	"log"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/redhat-nfvpe/smart-gateway/internal/pkg/incoming"
)

func TestCollectedTimeMetrics(t *testing.T) {
	c := incoming.NewInComing(incoming.COLLECTD)
	sample := c.GenerateSampleData("hostname", "pluginname")
	usetimestamp := true
	if collectd, ok := sample.(*incoming.Collectd); ok {
		metric, _ := NewCollectdMetric(usetimestamp, *collectd, 0)
		m := &dto.Metric{}
		metric.Write(m)
		ts := collectd.Time.Time()
		ms := ts.UnixNano() / 1000000
		log.Println(ms)
		if m.GetTimestampMs() != ms {
			t.Error("Time is not same as collected")
		}

	} else {
		t.Error("Empty sample string generated")
	}

}

func TestNonCollectedTimeMetrics(t *testing.T) {
	c := incoming.NewInComing(incoming.COLLECTD)
	sample := c.GenerateSampleData("hostname", "pluginname")

	usetimestamp := false
	if collectd, ok := sample.(*incoming.Collectd); ok {
		metric, _ := NewCollectdMetric(usetimestamp, *collectd, 0)
		m := &dto.Metric{}
		metric.Write(m)
		ts := collectd.Time.Time()
		ms := ts.UnixNano() / 1000000
		if m.GetTimestampMs() == ms {
			t.Error("Time is same as collected")
		}

	} else {
		t.Error("Empty sample string generated")
	}

}
