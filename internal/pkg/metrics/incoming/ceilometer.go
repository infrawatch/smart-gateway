package incoming

import (
	"time"
)

// CeilometerMetric struct represents metric data formated and sent by Ceilometer
type CeilometerMetric struct {
	Publisher string
	Timestamp time.Time
	EventType string
	//TODO(mmagr): include payload unmarshalled data here
}

/*************************** MetricDataFormat interface ****************************/

func (c *CeilometerMetric) GetName() string {
	return ""
}

func (c *CeilometerMetric) SetData(data MetricDataFormat) {}

func (c *CeilometerMetric) ParseInputJSON(json string) ([]MetricDataFormat, error) {
	return make([]MetricDataFormat, 0), nil
}

func (c *CeilometerMetric) GetKey() string {
	return ""
}

func (c *CeilometerMetric) GetItemKey() string {
	return ""
}

func (c *CeilometerMetric) ParseInputByte(data []byte) error {
	return nil
}

func (c *CeilometerMetric) GetInterval() float64 {
	return 0.0
}

func (c *CeilometerMetric) SetNew(new bool) {}

func (c *CeilometerMetric) ISNew() bool {
	return true
}

/*************************** tsdb.TSDB interface *****************************/
func (c *CeilometerMetric) GetLabels() map[string]string {
	return make(map[string]string)
}

func (c *CeilometerMetric) GetMetricName(index int) string {
	return ""
}

func (c *CeilometerMetric) GetMetricDesc(index int) string {
	return ""
}
