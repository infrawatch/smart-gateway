package incoming

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var (
	rexForPayload     = regexp.MustCompile(`\"payload\"\s*:\s*\[(.*)\]`)
	rexForOsloMessage = regexp.MustCompile(`"oslo.message"\s*:\s*"({.*})"`)
	rexForNestedQuote = regexp.MustCompile(`\\\"`)
)

const defaultCeilometerInterval = 5.0

// CeilometerMetric struct represents metric data formated and sent by Ceilometer
type CeilometerMetric struct {
	WithDataSource
	Publisher string                 `json:"publisher_id"`
	Payload   map[string]interface{} `json:"payload"`
	// analogy to collectd metric
	Plugin         string
	PluginInstance string
	Type           string
	TypeInstance   string
	Values         []float64
	new            bool
	wholeID        string
}

/*************************** MetricDataFormat interface ****************************/

func (c *CeilometerMetric) getwholeID() string {
	if c.wholeID == "" {
		if cnt, ok := c.Payload["counter_name"]; ok {
			c.wholeID = cnt.(string)
		} else {
			log.Printf("Did not find counter_name in metric payload: %v\n", c.Payload)
			c.wholeID = "unknown"
		}
	}
	return c.wholeID
}

//GetName returns name of Ceilometer "plugin" (analogically to CollectdMetric implementation)
func (c *CeilometerMetric) GetName() string {
	return c.Plugin
}

//GetValues returns Values. The purpose of this method is to be able to get metric Values
//from the interface object itself
func (c *CeilometerMetric) GetValues() []float64 {
	return c.Values
}

//SetData generates naming and value data analogicaly to CollectdMetric from counter data and resource_id
func (c *CeilometerMetric) SetData(data MetricDataFormat) {
	// example: counter_name=compute.instance.booting.time, resource_id=456
	// get Plugin -> compute
	plugParts := strings.Split(c.getwholeID(), ".")
	c.Plugin = plugParts[0]
	// get PluginInstance -> 456
	if resource, ok := c.Payload["resource_id"]; ok {
		c.PluginInstance = resource.(string)
	}
	// get Type -> instance
	if len(plugParts) > 1 {
		c.Type = plugParts[1]
	} else {
		c.Type = plugParts[0]
	}
	// get TypeInstance -> booting
	if len(plugParts) > 2 {
		c.TypeInstance = plugParts[2]
	}

	values := make([]float64, 0, 1)
	if val, ok := c.Payload["counter_volume"]; ok {
		values = append(values, val.(float64))
	} else {
		log.Printf("Did not find counter_volume in metric payload: %v\n", c.Payload)
	}
	c.Values = values
}

//sanitize search and removes all known issues in received data.
//TODO: Move this function to apputils
func (c *CeilometerMetric) sanitize(data string) string {
	sanitized := data
	// parse only relevant data
	sub := rexForOsloMessage.FindStringSubmatch(sanitized)
	if len(sub) == 2 {
		sanitized = rexForNestedQuote.ReplaceAllString(sub[1], `"`)
	} else {
		log.Printf("Failed to find oslo.message in given message: %s\n", data)
	}
	// avoid getting payload data wrapped in array
	item := rexForPayload.FindStringSubmatch(sanitized)
	if len(item) == 2 {
		sanitized = rexForPayload.ReplaceAllString(sanitized, fmt.Sprintf(`"payload":%s`, item[1]))
	}
	return sanitized
}

//ParseInputJSON ...
func (c *CeilometerMetric) ParseInputJSON(data string) ([]MetricDataFormat, error) {
	output := make([]MetricDataFormat, 0)
	sanitized := c.sanitize(data)
	// parse only relevant data
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal([]byte(sanitized), &c)
	if err != nil {
		return output, fmt.Errorf("error parsing json: %s", err)
	}
	c.SetNew(true)
	c.SetData(c)
	output = append(output, c)
	return output, nil
}

//GetKey ...
func (c *CeilometerMetric) GetKey() string {
	return c.Publisher
}

//GetItemKey returns name cache key analogically to CollectdMetric implementation
func (c *CeilometerMetric) GetItemKey() string {
	parts := []string{c.Plugin}
	if c.Plugin != c.Type {
		parts = append(parts, c.Type)
	}
	if c.PluginInstance != "" {
		parts = append(parts, c.PluginInstance)
	}
	if c.TypeInstance != "" {
		parts = append(parts, c.TypeInstance)
	}
	return strings.Join(parts, "_")
}

//ParseInputByte is not really used. It is here just to implement MetricDataFormat
//TODO: Remove this method from here and also CollectdMetric
func (c *CeilometerMetric) ParseInputByte(data []byte) error {
	_, err := c.ParseInputJSON(string(data))
	return err
}

//GetInterval returns hardcoded defaultCeilometerInterval, because Ceilometer metricDesc
//does not contain interval information (are not periodically sent at all) and any reasonable
//interval might be needed for expiry setting for Prometheus
//TODO: Make this configurable
func (c *CeilometerMetric) GetInterval() float64 {
	return defaultCeilometerInterval
}

//SetNew ...
func (c *CeilometerMetric) SetNew(new bool) {
	c.new = new
}

//ISNew ...
func (c *CeilometerMetric) ISNew() bool {
	return c.new
}

/*************************** tsdb.TSDB interface *****************************/

//GetLabels ...
func (c *CeilometerMetric) GetLabels() map[string]string {
	labels := make(map[string]string)
	if c.TypeInstance != "" {
		labels[c.Plugin] = c.TypeInstance
	} else {
		labels[c.Plugin] = c.PluginInstance
	}
	labels["publisher"] = c.Publisher
	if ctype, ok := c.Payload["counter_type"]; ok {
		labels["type"] = ctype.(string)
	} else {
		labels["type"] = "base"
	}
	if cproj, ok := c.Payload["project_id"]; ok {
		labels["project"] = cproj.(string)
	}
	if cres, ok := c.Payload["resource_id"]; ok {
		labels["resource"] = cres.(string)
	}
	if cunit, ok := c.Payload["counter_unit"]; ok {
		labels["unit"] = cunit.(string)
	}
	if cname, ok := c.Payload["counter_name"]; ok {
		labels["counter"] = cname.(string)
	}
	return labels
}

//GetMetricName ...
func (c *CeilometerMetric) GetMetricName(index int) string {
	nameParts := []string{"ceilometer", c.Plugin}
	if c.Plugin != c.Type {
		nameParts = append(nameParts, c.Type)
	}
	if c.TypeInstance != "" {
		nameParts = append(nameParts, c.TypeInstance)
	}
	return strings.Join(nameParts, "_")
}

//GetMetricDesc ...
func (c *CeilometerMetric) GetMetricDesc(index int) string {
	dstype := "counter"
	if ctype, ok := c.Payload["counter_type"]; ok {
		dstype = ctype.(string)
	}
	return fmt.Sprintf("Service Telemetry exporter: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		c.Plugin, c.Type, dstype, c.getwholeID())
}
