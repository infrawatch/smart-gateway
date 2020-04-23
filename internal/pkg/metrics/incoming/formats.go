package incoming

import (
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
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
	GetValues() []float64
}

type WithDataSource struct {
	DataSource saconfig.DataSource
}

//NewFromDataSource creates empty DataType according to given DataSource
func NewFromDataSource(source saconfig.DataSource) MetricDataFormat {
	switch source {
	case saconfig.DataSourceCollectd:
		return newCollectdMetric( /*...*/ )
	case saconfig.DataSourceCeilometer:
		return newCeilometerMetric()
	}
	return nil
}

//newCollectd  -- avoid calling this . Use factory method in incoming package
func newCollectdMetric() *CollectdMetric {
	metric := new(CollectdMetric)
	metric.DataSource = saconfig.DataSourceCollectd
	return metric
}

func newCeilometerMetric() *CeilometerMetric {
	metric := new(CeilometerMetric)
	metric.DataSource = saconfig.DataSourceCeilometer
	return metric
}

//ParseByte  parse incoming data
func ParseByte(dataItem MetricDataFormat, data []byte) error {
	return dataItem.ParseInputByte(data)
}
