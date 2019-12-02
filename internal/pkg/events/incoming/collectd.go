package incoming

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const GENERICINDEX = "collectd_generic"

var (
	rexForNestedQuote     = regexp.MustCompile(`\\\"`)
	rexForVes             = regexp.MustCompile(`"ves":"{(.*)}"`)
	collectdAlertSeverity = map[string]string{
		"OKAY":    "info",
		"WARNING": "warning",
		"FAILURE": "critical",
	}
)

type CollectdEvent struct {
	parsed    map[string]interface{}
	indexName string
}

//GetIndexName returns Elasticsearch index to which this event is or should be saved.
func (self *CollectdEvent) GetIndexName() string {
	if self.indexName == "" {
		result := GENERICINDEX
		if val, ok := self.parsed["labels"]; ok {
			switch rec := val.(type) {
			case map[string]interface{}:
				if value, ok := rec["alertname"].(string); ok {
					if index := strings.LastIndex(value, "_"); index > len("collectd_") {
						result = value[0:index]
					} else {
						result = value
					}
				}
			}
		}
		self.indexName = result
	}
	return self.indexName
}

//GetRawData returns sanitized and umarshalled event data.
func (self *CollectdEvent) GetRawData() interface{} {
	return self.parsed
}

//Sanitize search and removes all known issues in received data.
func (self *CollectdEvent) sanitize(jsondata string) string {
	// 1) if value for key "ves" is string, we convert it to json
	vesCleaned := jsondata
	sub := rexForVes.FindStringSubmatch(jsondata)
	if len(sub) == 2 {
		nested := rexForNestedQuote.ReplaceAllString(sub[1], `"`)
		vesCleaned = rexForVes.ReplaceAllString(jsondata, fmt.Sprintf(`"ves":{%s}`, nested))
	}
	// 2) event is received wrapped in array, so we remove it
	almostFinal := strings.TrimLeft(vesCleaned, "[")
	return strings.TrimRight(almostFinal, "]")
}

//ParseEvent sanitizes and unmarshals received event data.
func (self *CollectdEvent) ParseEvent(data string) error {
	sanitized := self.sanitize(data)
	err := json.Unmarshal([]byte(sanitized), &self.parsed)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

//assimilateMap recursively saves content of the given map to destination map of strings
func assimilateMap(theMap map[string]interface{}, destination *map[string]string) {
	for key, val := range theMap {
		switch value := val.(type) {
		case map[string]interface{}:
			// go one level deeper in the map
			assimilateMap(value, destination)
		case []interface{}:
			// transform slice value to comma separated list and assimilate it
			aList := make([]string, 0, len(value))
			for _, item := range value {
				if itm, ok := item.(string); ok {
					aList = append(aList, itm)
				}
			}
			(*destination)[key] = strings.Join(aList, ",")
		default:
			// assimilate KV pair
			(*destination)[key] = value.(string)
		}
	}
}

//GeneratePrometheusAlert generates alert PrometheusAlert from the event data
func (self *CollectdEvent) GeneratePrometheusAlert(generatorUrl string) PrometheusAlert {
	alert := PrometheusAlert{
		Labels:       make(map[string]string),
		Annotations:  make(map[string]string),
		GeneratorUrl: generatorUrl,
	}
	assimilateMap(self.parsed["labels"].(map[string]interface{}), &alert.Labels)
	assimilateMap(self.parsed["annotations"].(map[string]interface{}), &alert.Annotations)
	if value, ok := self.parsed["startsAt"].(string); ok {
		// ensure timestamps is in RFC3339
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, time.ANSIC} {
			stamp, err := time.Parse(layout, value)
			if err == nil {
				alert.StartsAt = stamp.Format(time.RFC3339)
				break
			}
		}
	}

	if value, ok := alert.Labels["severity"]; ok {
		if severity, ok := collectdAlertSeverity[value]; ok {
			alert.Labels["severity"] = severity
		}
	}
	alert.SetName()
	assimilateMap(self.parsed["annotations"].(map[string]interface{}), &alert.Labels)
	alert.SetSummary()

	alert.Labels["alertsource"] = "SmartGateway"
	return alert
}

//GeneratePrometheusAlert generates alert body for Prometheus Alert manager API
func (self *CollectdEvent) GeneratePrometheusAlertBody(generatorUrl string) ([]byte, error) {
	return json.Marshal(self.GeneratePrometheusAlert(generatorUrl))
}
