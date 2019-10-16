package tests

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"collectd.org/cdtime"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/incoming"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/stretchr/testify/assert"
)

type IncommingCollecdDataMatrix struct {
	Field    string
	Expected string
}

/*----------------------------- helper functions -----------------------------*/
//GenerateSampleCollectdData ...
func GenerateSampleCollectdData(hostname string, pluginname string) *incoming.Collectd {
	citfc := incoming.NewInComing(saconfig.DATA_TYPE_COLLECTD)
	collectd := citfc.(*incoming.Collectd)
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
func GetFieldStr(dataItem incoming.DataTypeInterface, field string) string {
	r := reflect.ValueOf(dataItem)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

/*----------------------------------------------------------------------------*/

func TestCollectdIncoming(t *testing.T) {
	emptySample := incoming.NewInComing(saconfig.DATA_TYPE_COLLECTD)
	sample := GenerateSampleCollectdData("hostname", "pluginname")
	jsonBytes, err := json.Marshal([]*incoming.Collectd{sample})
	if err != nil {
		t.Error("Failed to marshal incoming.Collectd to JSON")
	}
	jsonString := string(jsonBytes)

	t.Run("Test initialization of empty incoming.Collectd sample", func(t *testing.T) {
		assert.Emptyf(t, GetFieldStr(emptySample, "Plugin"), "Collectd data is not empty.")
		// test DSName behaviour
		if emptyCollectd, ok := emptySample.(*incoming.Collectd); ok {
			assert.Equal(t, "666", emptyCollectd.DSName(666))
			emptyCollectd.Values = []float64{1}
			assert.Equal(t, "value", emptyCollectd.DSName(666))
		} else {
			t.Errorf("Failed to convert empty incoming.DataTypeInterface to empty incoming.Collectd")
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
		esample := incoming.NewInComing(saconfig.DATA_TYPE_COLLECTD)
		errs := esample.ParseInputByte(jsonBytes)
		if errs == nil {
			sample3 := esample.(*incoming.Collectd)
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
		metricDesc := "Service Assurance exporter: 'pluginname' Type: 'collectd' Dstype: 'gauge' Dsname: 'value1'"
		assert.Equal(t, metricDesc, sample.GetMetricDesc(0))
		// test GetMetricName behaviour
		metricName := "sa_collectd_pluginname_collectd_value1"
		assert.Equal(t, metricName, sample.GetMetricName(0))
		sample.Type = sample.Plugin
		metricName = "sa_collectd_pluginname_value1"
		assert.Equal(t, metricName, sample.GetMetricName(0))
		sample.Dstypes = []string{"counter", "derive"}
		metricName1 := "sa_collectd_pluginname_value1_total"
		metricName2 := "sa_collectd_pluginname_value2_total"
		assert.Equal(t, metricName1, sample.GetMetricName(0))
		assert.Equal(t, metricName2, sample.GetMetricName(1))
	})
}
