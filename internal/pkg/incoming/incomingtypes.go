package incoming

//DataTypeInterface   ...
type DataTypeInterface interface {
	GetName() string
	SetData(data DataTypeInterface)
	ParseInputJSON(json string) error
	GetKey() string
	GetItemKey() string
	GenerateSampleData(key string, itemkey string) DataTypeInterface
	GenerateSampleJSON(key string, itemkey string) string
	ParseInputByte(data []byte) error
	//GenerateSamples(jsonstring string) *Interface
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

//GenerateData  Generates sample data in source format
func GenerateData(dataItem DataTypeInterface, key string, itemkey string) DataTypeInterface {
	return dataItem.GenerateSampleData(key, itemkey)
}

//GenerateJSON  Generates sample data  in json format
func GenerateJSON(dataItem DataTypeInterface, key string, itemkey string) string {
	return dataItem.GenerateSampleJSON(key, itemkey)
}

//ParseByte  parse incoming data
func ParseByte(dataItem DataTypeInterface, data []byte) error {
	return dataItem.ParseInputByte(data)
}
