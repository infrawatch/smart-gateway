package incoming

import (
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
)

//MetricDataFormat   ...
type MetricDataFormat interface {
	GetName() string
	SetData(data MetricDataFormat)
	ParseInputJSON(json string) ([]MetricDataFormat, error)
	GetKey() string
	GetItemKey() string
	ParseInputByte(data []byte) error
	GetInterval() float64
	SetNew(new bool)
	ISNew() bool
}

//NewFromDataSource creates empty DataType accorging to given DataSource
func NewFromDataSource(source saconfig.DataSource) MetricDataFormat {
	switch source {
	case saconfig.DATA_SOURCE_COLLECTD:
		return newCollectdMetric( /*...*/ )
	case saconfig.DATA_SOURCE_CEILOMETER:
		return newCeilometerMetric()
	}
	return nil
}

//newCollectd  -- avoid calling this . Use factory method in incoming package
func newCollectdMetric() *CollectdMetric {
	return new(CollectdMetric)
}

func newCeilometerMetric() *CeilometerMetric {
	return new(CeilometerMetric)
}

//ParseByte  parse incoming data
func ParseByte(dataItem MetricDataFormat, data []byte) error {
	return dataItem.ParseInputByte(data)
}
