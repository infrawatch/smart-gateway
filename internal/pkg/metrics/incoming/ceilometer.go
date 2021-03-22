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

// CeilometerMetric struct represents a single instance of metric data formated and sent by Ceilometer
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
	if ceilo, ok := data.(*CeilometerMetric); ok {
		c.Payload = ceilo.Payload
		c.Publisher = ceilo.Publisher

		plugParts := strings.Split(ceilo.getwholeID(), ".")
		c.Plugin = plugParts[0]
		// get PluginInstance -> 456
		if resource, ok := ceilo.Payload["resource_id"]; ok {
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
		if val, ok := ceilo.Payload["counter_volume"]; ok {
			values = append(values, val.(float64))
		} else {
			log.Printf("Did not find counter_volume in metric payload: %v\n", ceilo.Payload)
		}
		c.Values = values
		c.SetNew(true)
	}
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
		sanitized = rexForPayload.ReplaceAllString(sanitized, fmt.Sprintf(`"payload": [%s]`, strings.Join(item[1:], ",")))
	}
	return sanitized
}

//ParseInputJSON ... make this function type agnostic
func (c *CeilometerMetric) ParseInputJSON(data string) ([]MetricDataFormat, error) {
	dataPoints := make([]MetricDataFormat, 0)
	sanitized := c.sanitize(data)
	message := make(map[string]interface{})

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal([]byte(sanitized), &message)
	if err != nil {
		return nil, fmt.Errorf("error parsing json: %s", err)
	}

	for _, pl := range message["payload"].([]interface{}) {
		dP := CeilometerMetric{}
		if _, ok := message["publisher_id"].(string); !ok {
			return nil, fmt.Errorf("\"publisher_id\" not of type string")
		}
		if _, ok := pl.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("ceilometer metric payload not of type map[string]interface{}")
		}

		dP.Publisher = message["publisher_id"].(string)
		dP.Payload = pl.(map[string]interface{})

		dP.DataSource.SetFromString("ceilometer")
		dP.SetNew(true)
		dP.SetData(&dP)
		dataPoints = append(dataPoints, &dP)
	}

	return dataPoints, nil
}

//GetKey ...
func (c CeilometerMetric) GetKey() string {
	return c.Publisher
}

//GetItemKey returns name cache key analogically to CollectdMetric implementation
func (c *CeilometerMetric) GetItemKey() string {
	parts := []string{c.getwholeID()}
	if c.PluginInstance != "" {
		parts = append(parts, c.PluginInstance)
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
	if ctype, ok := c.Payload["counter_type"].(string); ok {
		labels["type"] = ctype
	} else {
		labels["type"] = "base"
	}
	if cproj, ok := c.Payload["project_id"].(string); ok {
		labels["project"] = cproj
	}
	if cres, ok := c.Payload["resource_id"].(string); ok {
		labels["resource"] = cres
	}
	if cunit, ok := c.Payload["counter_unit"].(string); ok {
		labels["unit"] = cunit
	}
	if cname, ok := c.Payload["counter_name"].(string); ok {
		labels["counter"] = cname
	}
	return labels
}

//GetMetricName ...
func (c *CeilometerMetric) GetMetricName(index int) string {
	nameParts := []string{"ceilometer"}
	cNameShards := strings.Split(c.getwholeID(), ".")
	nameParts = append(nameParts, cNameShards...)
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
