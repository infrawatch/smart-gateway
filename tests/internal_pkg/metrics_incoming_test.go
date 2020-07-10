package tests

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"collectd.org/cdtime"
	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/stretchr/testify/assert"
)

type IncommingCollecdDataMatrix struct {
	Field    string
	Expected string
}

const (
	ceilometerSampleMetricData = `{"request": {"oslo.version": "2.0", "oslo.message": "{\"message_id\": \"499e0dda-9298-4b03-a49c-d7affcedb6b9\", \"publisher_id\": \"telemetry.publisher.controller-0.redhat.local\", \"event_type\": \"metering\", \"priority\": \"SAMPLE\", \"payload\": [{\"source\": \"openstack\", \"counter_name\": \"disk.device.read.bytes\", \"counter_type\": \"cumulative\", \"counter_unit\": \"B\", \"counter_volume\": 18872832, \"user_id\": \"5df14d3577ff4c61b0837c268a8f4c70\", \"project_id\": \"5dfb98560ce74cf780c21fb18a5ad1de\", \"resource_id\": \"285778e1-c81b-427a-826a-ebb72467b665-vda\", \"timestamp\": \"2020-04-15T13:24:02.108816\", \"resource_metadata\": {\"display_name\": \"cirros\", \"name\": \"instance-00000001\", \"instance_id\": \"285778e1-c81b-427a-826a-ebb72467b665\", \"instance_type\": \"tiny\", \"host\": \"072f98fd91b8eec8d518aa8632f013438b587cee415dc944b39c5363\", \"instance_host\": \"compute-0.redhat.local\", \"flavor\": {\"id\": \"53e3164c-3dc0-4bd3-bb22-36a55aadd3fb\", \"name\": \"tiny\", \"vcpus\": 1, \"ram\": 256, \"disk\": 0, \"ephemeral\": 0, \"swap\": 0}, \"status\": \"active\", \"state\": \"running\", \"task_state\": \"\", \"image\": {\"id\": \"64f5d56e-e61d-43c1-af03-45b1faa89e99\"}, \"image_ref\": \"64f5d56e-e61d-43c1-af03-45b1faa89e99\", \"image_ref_url\": null, \"architecture\": \"x86_64\", \"os_type\": \"hvm\", \"vcpus\": 1, \"memory_mb\": 256, \"disk_gb\": 0, \"ephemeral_gb\": 0, \"root_gb\": 0, \"disk_name\": \"vda\"}, \"message_id\": \"5f312a0e-7f1c-11ea-a7a1-525400023f45\", \"monotonic_time\": null, \"message_signature\": \"8a47fa24471558f0af6963064e7ca1409237032c6c72a505f0acd51752f8f828\"}], \"timestamp\": \"2020-04-15 13:24:02.114844\"}"}}`
)

/*----------------------------- helper functions -----------------------------*/
//GenerateSampleCollectdData ...
func GenerateSampleCollectdData(hostname string, pluginname string) *incoming.CollectdMetric {
	citfc := incoming.NewFromDataSource(saconfig.DataSourceCollectd)
	collectd := citfc.(*incoming.CollectdMetric)
	collectd.Host = hostname
	collectd.Plugin = pluginname
	collectd.Type = "collectd"
	collectd.PluginInstance = "pluginnameinstance"
	collectd.Dstypes = []string{"gauge", "derive"}
	collectd.Dsnames = []string{"value1", "value2"}
	collectd.TypeInstance = "idle"
	collectd.Values = []float64{rand.Float64(), rand.Float64()}
	collectd.Time = cdtime.New(time.Now())
	return collectd
}

//GetFieldStr ...
func GetFieldStr(dataItem incoming.MetricDataFormat, field string) string {
	r := reflect.ValueOf(dataItem)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

/*----------------------------------------------------------------------------*/

func TestCollectdIncoming(t *testing.T) {
	emptySample := incoming.NewFromDataSource(saconfig.DataSourceCollectd)
	sample := GenerateSampleCollectdData("hostname", "pluginname")
	jsonBytes, err := json.Marshal([]*incoming.CollectdMetric{sample})
	if err != nil {
		t.Error("Failed to marshal incoming.Collectd to JSON")
	}
	jsonString := string(jsonBytes)

	t.Run("Test initialization of empty incoming.Collectd sample", func(t *testing.T) {
		assert.Emptyf(t, GetFieldStr(emptySample, "Plugin"), "Collectd data is not empty.")
		// test DSName behaviour
		if emptyCollectd, ok := emptySample.(*incoming.CollectdMetric); ok {
			assert.Equal(t, "666", emptyCollectd.DSName(666))
			emptyCollectd.Values = []float64{1}
			assert.Equal(t, "value", emptyCollectd.DSName(666))
		} else {
			t.Errorf("Failed to convert empty incoming.MetricDataFormat to empty incoming.CollectdMetric")
		}
		// test loading values from []byte and string
		_, errr := emptySample.ParseInputJSON("Error Json")
		assert.Error(t, errr, "Expected error got nil")
		data := []IncommingCollecdDataMatrix{
			{"Host", GetFieldStr(sample, "Host")},
			{"Plugin", GetFieldStr(sample, "Plugin")},
			{"Type", GetFieldStr(sample, "Type")},
			{"PluginInstance", GetFieldStr(sample, "PluginInstance")},
			{"Dstypes", GetFieldStr(sample, "Dstypes")},
			{"Dsnames", GetFieldStr(sample, "Dsnames")},
			{"TypeInstance", GetFieldStr(sample, "TypeInstance")},
			{"Values", GetFieldStr(sample, "Values")},
			{"Time", GetFieldStr(sample, "Time")},
		}
		sample2, errr := emptySample.ParseInputJSON(jsonString)
		if errr == nil {
			for _, testCase := range data {
				assert.Equal(t, testCase.Expected, GetFieldStr(sample2[0], testCase.Field))
			}
		} else {
			t.Errorf("Failed to initialize using ParseInputJSON: %s", err)
		}
		errr = emptySample.ParseInputByte([]byte("error string"))
		assert.Error(t, errr, "Expected error got nil")
		esample := incoming.NewFromDataSource(saconfig.DataSourceCollectd)
		errs := esample.ParseInputByte(jsonBytes)
		if errs == nil {
			sample3 := esample.(*incoming.CollectdMetric)
			for _, testCase := range data {
				assert.Equal(t, testCase.Expected, GetFieldStr(sample3, testCase.Field))
			}
		} else {
			t.Errorf("Failed to initialize using ParseInputByte: %s", err)
		}
	})

	t.Run("Test incoming.Collectd sample", func(t *testing.T) {
		assert.NotEmptyf(t, jsonBytes, "Empty sample string generated")
		// test DSName behaviour
		for index := range sample.Values {
			assert.Equal(t, fmt.Sprintf("value%d", index+1), sample.DSName(index))
		}
		assert.Equal(t, "pluginname", sample.GetName())
		// test GetItemKey behaviour
		assert.Equal(t, "pluginname_collectd_pluginnameinstance_idle", sample.GetItemKey())
		hold := sample.Type
		sample.Type = sample.Plugin
		assert.Equal(t, "pluginname_pluginnameinstance_idle", sample.GetItemKey())
		sample.Type = hold
		// test GetLabels behaviour
		labels := sample.GetLabels()
		assert.Contains(t, labels, "type")
		assert.Contains(t, labels, sample.Plugin)
		assert.Contains(t, labels, "instance")
		// test GetMetricDesc behaviour
		metricDesc := "Service Telemetry exporter: 'pluginname' Type: 'collectd' Dstype: 'gauge' Dsname: 'value1'"
		assert.Equal(t, metricDesc, sample.GetMetricDesc(0))
		// test GetMetricName behaviour
		metricName := "collectd_pluginname_collectd_value1"
		assert.Equal(t, metricName, sample.GetMetricName(0))
		sample.Type = sample.Plugin
		metricName = "collectd_pluginname_value1"
		assert.Equal(t, metricName, sample.GetMetricName(0))
		sample.Dstypes = []string{"counter", "derive"}
		metricName1 := "collectd_pluginname_value1_total"
		metricName2 := "collectd_pluginname_value2_total"
		assert.Equal(t, metricName1, sample.GetMetricName(0))
		assert.Equal(t, metricName2, sample.GetMetricName(1))
	})
}

func TestCeilometerIncoming(t *testing.T) {
	cm := incoming.NewFromDataSource(saconfig.DataSourceCeilometer)
	metric := cm.(*incoming.CeilometerMetric)

	t.Run("Test parsing of Ceilometer message", func(t *testing.T) {
		_, err := metric.ParseInputJSON(ceilometerSampleMetricData)
		if err != nil {
			t.Errorf("Ceilometer message parsing failed: %s\n", err)
		}
		// parameters
		assert.Equal(t, "telemetry.publisher.controller-0.redhat.local", metric.Publisher)
		assert.Equal(t, "disk", metric.Plugin)
		assert.Equal(t, "285778e1-c81b-427a-826a-ebb72467b665-vda", metric.PluginInstance)
		assert.Equal(t, "device", metric.Type)
		assert.Equal(t, "read", metric.TypeInstance)
		assert.Equal(t, []float64{float64(18872832)}, metric.Values)
		assert.Equal(t, saconfig.DataSourceCeilometer, metric.DataSource)
		// methods
		assert.Equal(t, 5.0, metric.GetInterval())
		assert.Equal(t, "disk_device_285778e1-c81b-427a-826a-ebb72467b665-vda_read", metric.GetItemKey())
		assert.Equal(t, "telemetry.publisher.controller-0.redhat.local", metric.GetKey())
		expectedLabels := map[string]string{
			"disk":      "read",
			"publisher": "telemetry.publisher.controller-0.redhat.local",
			"type":      "cumulative",
			"project":   "5dfb98560ce74cf780c21fb18a5ad1de",
			"resource":  "285778e1-c81b-427a-826a-ebb72467b665-vda",
			"unit":      "B",
			"counter":   "disk.device.read.bytes",
		}
		assert.Equal(t, expectedLabels, metric.GetLabels())
		assert.Equal(t, "Service Telemetry exporter: 'disk' Type: 'device' Dstype: 'cumulative' Dsname: 'disk.device.read.bytes'", metric.GetMetricDesc(0))
		assert.Equal(t, "ceilometer_disk_device_read", metric.GetMetricName(0))
		assert.Equal(t, "disk", metric.GetName())
		assert.Equal(t, []float64{float64(18872832)}, metric.GetValues())
		assert.Equal(t, true, metric.ISNew())
	})

}
