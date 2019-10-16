package incoming

import (
	"time"
)

type Ceilometer struct {
	Publisher string
	Timestamp time.Time
	EventType string
	//TODO(mmagr): include payload unmarshalled data here
}

func (c *Ceilometer) GetName() string {
	return ""
}

func (c *Ceilometer) SetData(data DataTypeInterface) {}

func (c *Ceilometer) ParseInputJSON(json string) ([]DataTypeInterface, error) {
	return make([]DataTypeInterface, 0), nil
}

func (c *Ceilometer) GetKey() string {
	return ""
}

func (c *Ceilometer) GetItemKey() string {
	return ""
}

func (c *Ceilometer) ParseInputByte(data []byte) error {
	return nil
}

func (c *Ceilometer) GetInterval() float64 {
	return 0.0
}

func (c *Ceilometer) SetNew(new bool) {}

func (c *Ceilometer) ISNew() bool {
	return true
}

func (c *Ceilometer) GetLabels() map[string]string {
	return make(map[string]string)
}

func (c *Ceilometer) GetMetricName(index int) string {
	return ""
}

func (c *Ceilometer) GetMetricDesc(index int) string {
	return ""
}
