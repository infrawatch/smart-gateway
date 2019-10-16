package incoming

import "github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"

//DataTypeInterface   ...
type DataTypeInterface interface {
	GetName() string
	SetData(data DataTypeInterface)
	ParseInputJSON(json string) ([]DataTypeInterface, error)
	GetKey() string
	GetItemKey() string
	ParseInputByte(data []byte) error
	GetInterval() float64
	SetNew(new bool)
	ISNew() bool
	TSDB
}

//TSDB  interface
type TSDB interface {
	//prometheus specifivreflect
	GetLabels() map[string]string
	GetMetricName(index int) string
	GetMetricDesc(index int) string
}

//NewInComing   ..
func NewInComing(t saconfig.DataType) DataTypeInterface {
	switch t {
	case saconfig.DATA_TYPE_COLLECTD:
		return newCollectd( /*...*/ )
	case saconfig.DATA_TYPE_CEILOMETER:
		return newCeilometer()
	}
	return nil
}

//newCollectd  -- avoid calling this . Use factory method in incoming package
func newCollectd() *Collectd {
	return new(Collectd)
}

func newCeilometer() *Ceilometer {
	return new(Ceilometer)
}

//ParseByte  parse incoming data
func ParseByte(dataItem DataTypeInterface, data []byte) error {
	return dataItem.ParseInputByte(data)
}
