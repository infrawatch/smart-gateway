package tests

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/stretchr/testify/assert"
)

const (
	EventsConfig = `
{
	"AMQP1Connections": [
		{"Url": "127.0.0.1:5672/collectd/notify", "DataSource": "collectd"},
		{"Url": "127.0.0.1:5672/ceilometer/events", "DataSource": "ceilometer"},
		{"Url": "127.0.0.1:5672/universal/events", "DataSource": "universal"}
	],
	"AMQP1EventURL": "127.0.0.1:5672/collectd/notify",
	"ElasticHostURL": "http://127.0.0.1:9200",
	"AlertManagerURL": "http://127.0.0.1:9093/api/v1/alerts",
	"ResetIndex": false,
	"Debug": true,
	"Prefetch": 101,
	"API":  {
	 "APIEndpointURL":  "http://127.0.0.1:8082",
	 "AMQP1PublishURL": "127.0.0.1:5672/collectd/alert"
   }

}
`
	MetricsConfig = `
{
	"AMQP1Connections": [
		{"Url": "127.0.0.1:5672/collectd/telemetry", "DataSource": "collectd"},
		{"Url": "127.0.0.1:5672/ceilometer/telemetry", "DataSource": "ceilometer"},
		{"Url": "127.0.0.1:5672/universal/telemetry", "DataSource": "universal"}
	],
	"AMQP1MetricURL": "127.0.0.1:5672/collectd/telemetry",
	"Exporterhost": "localhost",
	"Exporterport": 8081,
	"CPUStats": false,
	"DataCount": -1,
	"UseSample": false,
	"UseTimeStamp": true,
	"Debug": false,
	"Prefetch": 102,
	"Sample": {
		"HostCount": 10,
		"PluginCount": 100,
		"DataCount": -1
	}
}
`
)

/*----------------------------- helper functions -----------------------------*/

//GenerateSampleCacheData  ....
func GenerateTestConfig(content string) (string, error) {
	file, err := ioutil.TempFile(".", "smart_gateway_config_test")
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.WriteString(content)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

/*----------------------------------------------------------------------------*/

type ConfigDataMatrix struct {
	Field string
	Value interface{}
}

type ConfigDataTestRun struct {
	Name    string
	Content string
	Loader  string
	Matrix  []ConfigDataMatrix
}

func TestUnstructuredData(t *testing.T) {
	testRuns := []ConfigDataTestRun{
		{
			Name:    "Test events config",
			Content: EventsConfig,
			Loader:  "event",
			Matrix: []ConfigDataMatrix{
				{"AMQP1EventURL", "127.0.0.1:5672/collectd/notify"},
				{"ElasticHostURL", "http://127.0.0.1:9200"},
				{"AlertManagerURL", "http://127.0.0.1:9093/api/v1/alerts"},
				{"ResetIndex", false},
				{"Debug", true},
				{"Prefetch", 101},
			},
		},
		{
			Name:    "Test metrics config",
			Content: MetricsConfig,
			Loader:  "metric",
			Matrix: []ConfigDataMatrix{
				{"AMQP1MetricURL", "127.0.0.1:5672/collectd/telemetry"},
				{"Exporterhost", "localhost"},
				{"Exporterport", 8081},
				{"CPUStats", false},
				{"DataCount", -1},
				{"UseTimeStamp", true},
				{"Debug", false},
				{"Prefetch", 102},
			},
		},
	}

	for _, run := range testRuns {
		confPath, err := GenerateTestConfig(run.Content)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(confPath)
		cfg, err := saconfig.LoadConfiguration(confPath, run.Loader)
		if err != nil {
			t.Fatal(err)
		}
		if run.Loader == "metric" {
			cfg = cfg.(*saconfig.MetricConfiguration)
		} else {
			cfg = cfg.(*saconfig.EventConfiguration)
		}
		t.Run(run.Name, func(t *testing.T) {
			reflectedVal := reflect.ValueOf(cfg)
			for _, testCase := range run.Matrix {
				field := reflectedVal.Elem().FieldByName(testCase.Field)
				if !field.IsValid() {
					t.Fail()
				}
				value := reflect.ValueOf(testCase.Value)
				if !field.IsValid() {
					t.Errorf("Failed to parse field %s.", testCase.Field)
					continue
				}
				switch value.Type().String() {
				case "bool":
					assert.Equal(t, value.Bool(), field.Bool())
				case "int":
					assert.Equal(t, value.Int(), field.Int())
				default:
					assert.Equal(t, value.String(), field.String())
				}
			}
		})
	}
}

func TestStructuredData(t *testing.T) {
	confPath, err := GenerateTestConfig(EventsConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(confPath)
	cfg, err := saconfig.LoadConfiguration(confPath, "event")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test structured data of event config", func(t *testing.T) {
		apiStruct := saconfig.EventAPIConfig{APIEndpointURL: "http://127.0.0.1:8082", AMQP1PublishURL: "127.0.0.1:5672/collectd/alert"}
		assert.Equal(t, apiStruct, cfg.(*saconfig.EventConfiguration).API)
	})

	t.Run("Test structured AMQP connections", func(t *testing.T) {
		connStruct := []saconfig.AMQPConnection{
			saconfig.AMQPConnection{Url: "127.0.0.1:5672/collectd/notify", DataSource: "collectd", DataSourceId: 1},
			saconfig.AMQPConnection{Url: "127.0.0.1:5672/ceilometer/events", DataSource: "ceilometer", DataSourceId: 2},
			saconfig.AMQPConnection{Url: "127.0.0.1:5672/universal/events", DataSource: "universal", DataSourceId: 0},
		}
		assert.Equal(t, connStruct, cfg.(*saconfig.EventConfiguration).AMQP1Connections)
	})
}
