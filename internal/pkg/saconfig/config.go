package saconfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//EventConfiguration  ...
type EventConfiguration struct {
	Debug               bool           `json:"Debug"`
	AMQP1EventURL       string         `json:"AMQP1EventURL"`
	ElasticHostURL      string         `json:"ElasticHostURL"`
	API                 EventAPIConfig `json:"API"`
	AlertManagerURL     string         `json:"AlertManagerURL"`
	AlertManagerEnabled bool           `json:"AlertManagerEnabled"`
	APIEnabled          bool           `json:"APIEnabled"`
	PublishEventEnabled bool           `json:"PublishEventEnabled"`
	ResetIndex          bool           `json:"ResetIndex"`
	Prefetch            int            `json:"Prefetch"`
	UniqueName          string         `json:"UniqueName"`
	ServiceType         string         `json:"ServiceType"`
	IgnoreString        string         `json:"-"` //TODO(mmagr): ?
}

//EventAPIConfig ...
type EventAPIConfig struct {
	APIEndpointURL  string `json:"APIEndpointURL"`  //API endpoint
	AMQP1PublishURL string `json:"AMQP1PublishURL"` // new amqp address to send notifications
}

//MetricConfiguration   ....
type MetricConfiguration struct {
	Debug          bool             `json:"Debug"`
	TestServer     bool             `json:"TestServer"`
	AMQP1MetricURL string           `json:"AMQP1MetricURL"`
	CPUStats       bool             `json:"CPUStats"`
	Exporterhost   string           `json:"Exporterhost"`
	Exporterport   int              `json:"Exporterport"`
	Prefetch       int              `json:"Prefetch"`
	DataCount      int              `json:"DataCount"` //-1 for ever which is default //TODO(mmagr): config implementation does not have a way to for default value, implement one?
	UseSample      bool             `json:"UseSample"`
	UseTimeStamp   bool             `json:"UseTimeStamp"`
	Sample         SampleDataConfig `json:"Sample"`
	UniqueName     string           `json:"UniqueName"`
	ServiceType    string           `json:"ServiceType"`
	IgnoreString   string           `json:"-"` //TODO(mmagr): ?
}

//SampleDataConfig  ...
type SampleDataConfig struct {
	HostCount   int `json:"HostCount"`   //no of host to simulate
	PluginCount int `json:"PluginCount"` //No of plugin count per hosts
	DataCount   int `json:"DataCount"`   //-1 for ever howmany times sample data should be generated
}

//TODO(mmagr): aaply DRY principle here

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
