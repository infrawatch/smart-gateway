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

type LoadFunc func(p []byte) (n int, err error)

type ConfigDataTestRun struct {
	Name    string
	Content string
	Loader  string
	Matrix  []ConfigDataMatrix
}

func TestUnstructuredData(t *testing.T) {
	testRuns := []ConfigDataTestRun{
		ConfigDataTestRun{
			Name:    "Test events config",
			Content: EventsConfig,
			Loader:  "LoadEventConfig",
			Matrix: []ConfigDataMatrix{
				ConfigDataMatrix{"AMQP1EventURL", "127.0.0.1:5672/collectd/notify"},
				ConfigDataMatrix{"ElasticHostURL", "http://127.0.0.1:9200"},
				ConfigDataMatrix{"AlertManagerURL", "http://127.0.0.1:9093/api/v1/alerts"},
				ConfigDataMatrix{"ResetIndex", false},
				ConfigDataMatrix{"Debug", true},
				ConfigDataMatrix{"Prefetch", 101},
			},
		},
		ConfigDataTestRun{
			Name:    "Test metrics config",
			Content: MetricsConfig,
			Loader:  "LoadMetricConfig",
			Matrix: []ConfigDataMatrix{
				ConfigDataMatrix{"AMQP1MetricURL", "127.0.0.1:5672/collectd/telemetry"},
				ConfigDataMatrix{"Exporterhost", "localhost"},
				ConfigDataMatrix{"Exporterport", 8081},
				ConfigDataMatrix{"CPUStats", false},
				ConfigDataMatrix{"DataCount", -1},
				ConfigDataMatrix{"UseTimeStamp", true},
				ConfigDataMatrix{"Debug", false},
				ConfigDataMatrix{"Prefetch", 102},
			},
		},
	}

	for _, run := range testRuns {
		confPath, err := GenerateTestConfig(run.Content)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(confPath)
		var cfg interface{}
		switch run.Loader {
		case "LoadEventConfig":
			cfg = saconfig.LoadEventConfig(confPath)
		case "LoadMetricConfig":
			cfg = saconfig.LoadMetricConfig(confPath)
		default:
			t.Errorf("Unknown config loader %v", run.Loader)
		}
		t.Run(run.Name, func(t *testing.T) {
			reflectedVal := reflect.ValueOf(cfg)
			for _, testCase := range run.Matrix {
				field := reflectedVal.FieldByName(testCase.Field)
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
	t.Run("Test structured data of event config", func(t *testing.T) {
		confPath, err := GenerateTestConfig(EventsConfig)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(confPath)
		cfg := saconfig.LoadEventConfig(confPath)

		apiStruct := saconfig.EventAPIConfig{APIEndpointURL: "http://127.0.0.1:8082", AMQP1PublishURL: "127.0.0.1:5672/collectd/alert"}
		assert.Equal(t, apiStruct, cfg.API)
	})
}
