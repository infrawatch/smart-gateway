package tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"collectd.org/cdtime"
	"github.com/infrawatch/smart-gateway/internal/pkg/metrics/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

type IncommingCollecdDataMatrix struct {
	Field    string
	Expected string
}

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

//CeilometerMetricTemplate holds correct parsings for comparing against parsed results
type CeilometerMetricTestTemplate struct {
	TestInput        jsoniter.RawMessage `json:"testInput"`
	ValidatedResults []*struct {
		Publisher      string            `json:"publisher"`
		Plugin         string            `json:"plugin"`
		PluginInstance string            `json:"plugin_instance"`
		Type           string            `json:"type"`
		TypeInstance   string            `json:"type_instance"`
		Name           string            `json:"name"`
		Key            string            `json:"key"`
		ItemKey        string            `json:"item_Key"`
		Description    string            `json:"description"`
		MetricName     string            `json:"metric_name"`
		Labels         map[string]string `json:"labels"`
		Values         []float64         `json:"values"`
		ISNew          bool
		Interval       float64
		DataSource     saconfig.DataSource
	} `json:"validatedResults"`
}

func CeilometerMetricTestTemplateFromJSON(jsonData string) (*CeilometerMetricTestTemplate, error) {
	var testData CeilometerMetricTestTemplate
	err := json.Unmarshal([]byte(jsonData), &testData)
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %s", err)
	}

	for _, r := range testData.ValidatedResults {
		r.Interval = 5.0
		r.DataSource = saconfig.DataSourceCeilometer
		r.ISNew = true
	}
	return &testData, nil
}

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

	var tests = make(map[string]jsoniter.RawMessage)

	testDataJSON, err := ioutil.ReadFile("messages/metric-tests.json")
	if err != nil {
		t.Errorf("Failed loading test data: %s\n", err)
	}

	err = json.Unmarshal([]byte(testDataJSON), &tests)
	if err != nil {
		t.Errorf("error parsing json: %s", err)
	}

	t.Run("Test parsing of Ceilometer message", func(t *testing.T) {
		testData, err := CeilometerMetricTestTemplateFromJSON(string(tests["CeilometerMetrics"]))
		if err != nil {
			t.Errorf("Failed loading ceilometer metric test data: %s", err)
		}
		metrics, err := metric.ParseInputJSON(string(testData.TestInput))
		if err != nil {
			t.Errorf("Ceilometer message parsing failed: %s\n", err)
		}

		for index, standard := range testData.ValidatedResults {
			if m, ok := metrics[index].(*incoming.CeilometerMetric); ok {
				assert.Equal(t, standard.Publisher, m.Publisher)
				assert.Equal(t, standard.Plugin, m.Plugin)
				assert.Equal(t, standard.PluginInstance, m.PluginInstance)
				assert.Equal(t, standard.Type, m.Type)
				assert.Equal(t, standard.TypeInstance, m.TypeInstance)
				assert.Equal(t, standard.Values, m.GetValues())
				assert.Equal(t, standard.DataSource, m.DataSource)
				assert.Equal(t, standard.Interval, m.GetInterval())
				assert.Equal(t, standard.ItemKey, m.GetItemKey())
				assert.Equal(t, standard.Key, m.GetKey())
				assert.Equal(t, standard.Labels, m.GetLabels())
				assert.Equal(t, standard.Name, m.GetName())
				assert.Equal(t, standard.ISNew, m.ISNew())
				assert.Equal(t, standard.Description, m.GetMetricDesc(0))
				assert.Equal(t, standard.MetricName, m.GetMetricName(0))
			}
		}
	})
}

func TestCeilometerGetLabels(t *testing.T) {
	// test data
	cm := incoming.CeilometerMetric{
		Payload: map[string]interface{}{},
	}

	t.Run("Missing fields", func(t *testing.T) {
		cm.GetLabels()
	})

	t.Run("Wrong field type", func(t *testing.T) {
		cm.Payload["project_id"] = nil
		cm.Payload["resource_id"] = nil
		cm.Payload["counter_unit"] = nil
		cm.Payload["counter_name"] = nil
		cm.GetLabels()
	})
}
