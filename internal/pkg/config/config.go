package saconfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//EventConfiguration  ...
type EventConfiguration struct {
	Debug               bool
	AMQP1EventURL       string
	ElasticHostURL      string
	API                 EventAPIConfig
	AlertManagerURL     string
	AlertManagerEnabled bool
	APIEnabled          bool
	PublishEventEnabled bool
	ResetIndex          bool
	IgnoreString        string `json:"-"`
}

//EventAPIConfig ...
type EventAPIConfig struct {
	APIEndpointURL  string //API endpoint
	AMQP1PublishURL string // new amqp address to send notifications
}

//MetricConfiguration   ....
type MetricConfiguration struct {
	Debug          bool
	AMQP1MetricURL string
	CPUStats       bool
	Exporterhost   string
	Exporterport   int
	DataCount      int //-1 for ever which is default
	UseSample      bool
	Sample         SampleDataConfig
	IgnoreString   string `json:"-"`
}

//SampleDataConfig  ...
type SampleDataConfig struct {
	HostCount   int //no of host to simulate
	PluginCount int //No of plugin count per hosts
	DataCount   int //-1 for ever howmany times sample data should be generated

}

//LoadMetricConfig ....
func LoadMetricConfig(path string) MetricConfiguration {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing. ", err)
	}
	var config MetricConfiguration

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Config Parse Error: ", err)
	}

	return config
}

//LoadEventConfig ....
func LoadEventConfig(path string) EventConfiguration {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing. ", err)
	}
	var config EventConfiguration
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Config Parse Error: ", err)
	}

	return config
}
