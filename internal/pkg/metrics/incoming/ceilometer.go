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

//GetName ...
func (c *CeilometerMetric) GetName() string {
	return ""
}

//SetData ...
func (c *CeilometerMetric) SetData(data MetricDataFormat) {}

//ParseInputJSON ...
func (c *CeilometerMetric) ParseInputJSON(json string) ([]MetricDataFormat, error) {
	return make([]MetricDataFormat, 0), nil
}

//GetKey ...
func (c *CeilometerMetric) GetKey() string {
	return ""
}

//GetItemKey ...
func (c *CeilometerMetric) GetItemKey() string {
	return ""
}

//ParseInputByte ...
func (c *CeilometerMetric) ParseInputByte(data []byte) error {
	return nil
}

//GetInterval ...
func (c *CeilometerMetric) GetInterval() float64 {
	return 0.0
}

//SetNew ...
func (c *CeilometerMetric) SetNew(new bool) {}

//ISNew ...
func (c *CeilometerMetric) ISNew() bool {
	return true
}

/*************************** tsdb.TSDB interface *****************************/

//GetLabels ...
func (c *CeilometerMetric) GetLabels() map[string]string {
	return make(map[string]string)
}

//GetMetricName ...
func (c *CeilometerMetric) GetMetricName(index int) string {
	return ""
}

//GetMetricDesc ...
func (c *CeilometerMetric) GetMetricDesc(index int) string {
	return ""
}
