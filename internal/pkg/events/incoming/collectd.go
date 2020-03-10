package incoming

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

//collectdGenericIndex value represents ElasticSearch index name for data from which it
// is not possible to clearly construct indexs name
const collectdGenericIndex = "collectd_generic"

var (
	rexForNestedQuote     = regexp.MustCompile(`\\\"`)
	rexForVes             = regexp.MustCompile(`"ves":"{(.*)}"`)
	collectdAlertSeverity = map[string]string{
		"OKAY":    "info",
		"WARNING": "warning",
		"FAILURE": "critical",
	}
)

//CollectdEvent implements EventDataFormat interface and holds event message data from collectd.
type CollectdEvent struct {
	sanitized string
	parsed    map[string]interface{}
	indexName string
}

//GetIndexName returns Elasticsearch index to which this event is or should be saved.
func (evt *CollectdEvent) GetIndexName() string {
	if evt.indexName == "" {
		result := collectdGenericIndex
		if val, ok := evt.parsed["labels"]; ok {
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
		evt.indexName = result
	}
	return evt.indexName
}

//GetRawData returns sanitized and umarshalled event data.
func (evt *CollectdEvent) GetRawData() interface{} {
	return evt.parsed
}

//GetSanitized returns sanitized event data
func (evt *CollectdEvent) GetSanitized() string {
	return evt.sanitized
}

//sanitize search and removes all known issues in received data.
func (evt *CollectdEvent) sanitize(jsondata string) string {
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
func (evt *CollectdEvent) ParseEvent(data string) error {
	evt.sanitized = evt.sanitize(data)
	err := json.Unmarshal([]byte(evt.sanitized), &evt.parsed)
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

//GeneratePrometheusAlert generates PrometheusAlert from the event data
func (evt *CollectdEvent) GeneratePrometheusAlert(generatorURL string) PrometheusAlert {
	alert := PrometheusAlert{
		Labels:       make(map[string]string),
		Annotations:  make(map[string]string),
		GeneratorURL: generatorURL,
	}
	assimilateMap(evt.parsed["labels"].(map[string]interface{}), &alert.Labels)
	assimilateMap(evt.parsed["annotations"].(map[string]interface{}), &alert.Annotations)
	if value, ok := evt.parsed["startsAt"].(string); ok {
		// ensure timestamps is in RFC3339
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, time.ANSIC, isoTimeLayout} {
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
		} else {
			alert.Labels["severity"] = unknownSeverity
		}
	} else {
		alert.Labels["severity"] = unknownSeverity
	}

	alert.SetName()
	assimilateMap(evt.parsed["annotations"].(map[string]interface{}), &alert.Labels)
	alert.SetSummary()

	alert.Labels["alertsource"] = "SmartGateway"
	return alert
}

//GeneratePrometheusAlertBody generates alert body for Prometheus Alert manager API
func (evt *CollectdEvent) GeneratePrometheusAlertBody(generatorURL string) ([]byte, error) {
	return json.Marshal(evt.GeneratePrometheusAlert(generatorURL))
}
