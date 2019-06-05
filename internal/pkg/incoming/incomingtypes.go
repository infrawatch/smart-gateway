package incoming

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

//DataType   ..
type DataType int

//COLLECTD
const (
	COLLECTD DataType = 1 << iota
)

//NewInComing   ..
func NewInComing(t DataType) DataTypeInterface {
	switch t {
	case COLLECTD:
		return newCollectd( /*...*/ )
	}
	return nil
}

//newCollectd  -- avoid calling this . Use factory method in incoming package
func newCollectd() *Collectd {
	return new(Collectd)
}

//ParseByte  parse incoming data
func ParseByte(dataItem DataTypeInterface, data []byte) error {
	return dataItem.ParseInputByte(data)
}
