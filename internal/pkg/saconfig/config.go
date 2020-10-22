package saconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

/************************** DataSource implementation **************************/

//DataSource indentifies a format of incoming data in the message bus channel.
type DataSource int

const (
	//DataSourceUniversal marks all types of data sorces which send data in smart-gateway universal format
	DataSourceUniversal DataSource = iota
	//DataSourceCollectd marks collectd as data source for metrics and(or) events
	DataSourceCollectd
	//DataSourceCeilometer marks Ceilometer as data source for metrics and(or) events
	DataSourceCeilometer
)

//String returns human readable data type identification.
func (src DataSource) String() string {
	return []string{"universal", "collectd", "ceilometer"}[src]
}

//SetFromString resets value according to given human readable identification. Returns false if invalid identification was given.
func (src *DataSource) SetFromString(name string) bool {
	for index, value := range []string{"universal", "collectd", "ceilometer"} {
		if name == value {
			*src = DataSource(index)
			return true
		}
	}
	return false
}

/*********************** AMQPConnection implementation ***********************/

//AMQPConnection identifies single messagebus connection and expected format of incoming data.
type AMQPConnection struct {
	URL          string `json:"URL"`
	DataSource   string `json:"DataSource"`
	DataSourceID DataSource
}

/********************* EventConfiguration implementation *********************/

//EventAPIConfig ...
type EventAPIConfig struct {
	APIEndpointURL  string `json:"APIEndpointURL"`  //API endpoint
	AMQP1PublishURL string `json:"AMQP1PublishURL"` // new amqp address to send notifications
}

type HandlerPath struct {
	Path       string `json:"Path"`
	DataSource string `json:"DataSource"`
}

//EventConfiguration ...
type EventConfiguration struct {
	Debug               bool             `json:"Debug"`
	AMQP1EventURL       string           `json:"AMQP1EventURL"`
	AMQP1Connections    []AMQPConnection `json:"AMQP1Connections"`
	ElasticHostURL      string           `json:"ElasticHostURL"`
	UseBasicAuth        bool             `json:"UseBasicAuth"`
	ElasticUser         string           `json:"ElasticUser"`
	ElasticPass         string           `json:"ElasticPass"`
	API                 EventAPIConfig   `json:"API"`
	AlertManagerURL     string           `json:"AlertManagerURL"`
	AlertManagerEnabled bool             `json:"AlertManagerEnabled"`
	APIEnabled          bool             `json:"APIEnabled"`
	PublishEventEnabled bool             `json:"PublishEventEnabled"`
	ResetIndex          bool             `json:"ResetIndex"`
	Prefetch            int              `json:"Prefetch"`
	UniqueName          string           `json:"UniqueName"`
	ServiceType         string           `json:"ServiceType"`
	IgnoreString        string           `json:"-"` //TODO(mmagr): ?
	UseTLS              bool             `json:"UseTls"`
	TLSServerName       string           `json:"TlsServerName"`
	TLSClientCert       string           `json:"TlsClientCert"`
	TLSClientKey        string           `json:"TlsClientKey"`
	TLSCaCert           string           `json:"TlsCaCert"`
	HandlerPlugins      []HandlerPath    `json:"HandlerPlugin"`
}

/******************** MetricConfiguration implementation *********************/

//MetricConfiguration ...
type MetricConfiguration struct {
	Debug            bool             `json:"Debug"`
	AMQP1MetricURL   string           `json:"AMQP1MetricURL"`
	AMQP1Connections []AMQPConnection `json:"AMQP1Connections"`
	CPUStats         bool             `json:"CPUStats"`
	Exporterhost     string           `json:"Exporterhost"`
	Exporterport     int              `json:"Exporterport"`
	Prefetch         int              `json:"Prefetch"`
	DataCount        int              `json:"DataCount"` //-1 for ever which is default //TODO(mmagr): config implementation does not have a way to for default value, implement one?
	UseTimeStamp     bool             `json:"UseTimeStamp"`
	UniqueName       string           `json:"UniqueName"`
	ServiceType      string           `json:"ServiceType"`
	IgnoreString     string           `json:"-"` //TODO(mmagr): ?
}

/*****************************************************************************/

//LoadConfiguration loads and unmarshals configuration file by given path and type
func LoadConfiguration(path string, confType string) (interface{}, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing.", err)
	}

	var config interface{}
	switch confType {
	case "metric":
		config = new(MetricConfiguration)
	case "event":
		config = new(EventConfiguration)
	}
	err = json.Unmarshal(file, &config)

	var connections []AMQPConnection
	switch confType {
	case "metric":
		connections = config.(*MetricConfiguration).AMQP1Connections
	case "event":
		connections = config.(*EventConfiguration).AMQP1Connections
	}
	for index, conn := range connections {
		var dts *DataSource
		switch confType {
		case "metric":
			dts = &config.(*MetricConfiguration).AMQP1Connections[index].DataSourceID
		case "event":
			dts = &config.(*EventConfiguration).AMQP1Connections[index].DataSourceID
		}
		if ok := dts.SetFromString(conn.DataSource); !ok {
			err = fmt.Errorf("invalid AMQP connection data source '%s'", conn.DataSource)
		}
	}
	return config, err
}
