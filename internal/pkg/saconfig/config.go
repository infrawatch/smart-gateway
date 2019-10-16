package saconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

/************************** DataType implementation **************************/

//DataType indentifies a format of incoming data in the message bus channel.
type DataType int

const (
	DATA_TYPE_UNIVERSAL DataType = iota
	DATA_TYPE_COLLECTD
	DATA_TYPE_CEILOMETER
)

//String returns human readable data type identification.
func (self *DataType) String() string {
	return []string{"universal", "collectd", "ceilometer"}[*self]
}

//SetFromString resets value according to given human readable identification. Returns false if invalid identification was given.
func (self *DataType) SetFromString(name string) bool {
	for index, value := range []string{"universal", "collectd", "ceilometer"} {
		if name == value {
			*self = DataType(index)
			return true
		}
	}
	return false
}

/*********************** AMQPConnection implementation ***********************/

//AMQPConnection identifies single messagebus connection and expected format of incoming data.
type AMQPConnection struct {
	Url        string `json:"Url"`
	DataType   string `json:"DataType"`
	DataTypeId DataType
}

/********************* EventConfiguration implementation *********************/

//EventAPIConfig ...
type EventAPIConfig struct {
	APIEndpointURL  string `json:"APIEndpointURL"`  //API endpoint
	AMQP1PublishURL string `json:"AMQP1PublishURL"` // new amqp address to send notifications
}

//EventConfiguration ...
type EventConfiguration struct {
	Debug               bool             `json:"Debug"`
	AMQP1EventURL       string           `json:"AMQP1EventURL"`
	AMQP1Connections    []AMQPConnection `json:"AMQP1Connections"`
	ElasticHostURL      string           `json:"ElasticHostURL"`
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

//LoadConfig loads and unmarshals configuration file by given path and type
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
		var dti *DataType
		switch confType {
		case "metric":
			dti = &config.(*MetricConfiguration).AMQP1Connections[index].DataTypeId
		case "event":
			dti = &config.(*EventConfiguration).AMQP1Connections[index].DataTypeId
		}
		if ok := dti.SetFromString(conn.DataType); !ok {
			err = fmt.Errorf("Invalid AMQP connection data type %s", conn.DataType)
		}
	}
	return config, err
}
