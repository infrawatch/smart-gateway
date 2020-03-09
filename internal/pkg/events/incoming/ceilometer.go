package incoming

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

//ceilomterGenericIndex value represents ElasticSearch index name for data from which it
// is not possible to clearly construct indexs name
const ceilometerGenericIndex = "ceilometer_generic"

var (
	rexForPayload           = regexp.MustCompile(`\"payload\"\s*:\s*\[(.*)\]`)
	rexForOsloMessage       = regexp.MustCompile(`"oslo.message"\s*:\s*"({.*})"`)
	ceilometerAlertSeverity = map[string]string{
		"audit":    "info",
		"info":     "info",
		"sample":   "info",
		"warn":     "warning",
		"warning":  "warning",
		"critical": "critical",
		"error":    "critical",
		"AUDIT":    "info",
		"INFO":     "info",
		"SAMPLE":   "info",
		"WARN":     "warning",
		"WARNING":  "warning",
		"CRITICAL": "critical",
		"ERROR":    "critical",
	}
)

//AlertKeySurrogate translates case for fields for AlertManager
type AlertKeySurrogate struct {
	Parsed string
	Label  string
}

//CeilometerEvent implements EventDataFormat interface and holds event message data from collectd.
type CeilometerEvent struct {
	sanitized string
	parsed    map[string]interface{}
	indexName string
}

//GetIndexName returns Elasticsearch index to which this event is or should be saved.
func (evt *CeilometerEvent) GetIndexName() string {
	if evt.indexName == "" {
		result := ceilometerGenericIndex
		if val, ok := evt.parsed["event_type"]; ok {
			switch rec := val.(type) {
			case string:
				parts := strings.Split(rec, ".")
				if len(parts) > 1 {
					result = strings.ReplaceAll(strings.Join(parts[:len(parts)-1], "_"), "-", "_")
				} else {
					result = fmt.Sprintf("%v", parts[0])
				}
				if !strings.HasPrefix(result, "ceilometer_") {
					result = fmt.Sprintf("ceilometer_%s", result)
				}
			}
		}
		evt.indexName = result
	}
	return evt.indexName
}

//GetRawData returns sanitized and umarshalled event data.
func (evt *CeilometerEvent) GetRawData() interface{} {
	return evt.parsed
}

//GetSanitized returns sanitized event data
func (evt *CeilometerEvent) GetSanitized() string {
	return evt.sanitized
}

//sanitize search and removes all known issues in received data.
func (evt *CeilometerEvent) sanitize(jsondata string) string {
	sanitized := jsondata
	// parse only relevant data
	sub := rexForOsloMessage.FindStringSubmatch(sanitized)
	if len(sub) == 2 {
		sanitized = rexForNestedQuote.ReplaceAllString(sub[1], `"`)
	} else {
		log.Printf("Failed to find oslo.message in Ceilometer event: %s\n", jsondata)
	}
	// avoid getting payload data wrapped in array
	item := rexForPayload.FindStringSubmatch(sanitized)
	if len(item) == 2 {
		sanitized = rexForPayload.ReplaceAllString(sanitized, fmt.Sprintf(`"payload":%s`, item[1]))
	}
	return sanitized
}

//ParseEvent sanitizes and unmarshals received event data.
func (evt *CeilometerEvent) ParseEvent(data string) error {
	evt.sanitized = evt.sanitize(data)
	err := json.Unmarshal([]byte(evt.sanitized), &evt.parsed)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// transforms traits key into map[string]interface{}
	if payload, ok := evt.parsed["payload"]; ok {
		newPayload := make(map[string]interface{})
		if typedPayload, ok := payload.(map[string]interface{}); ok {
			if traitData, ok := typedPayload["traits"]; ok {
				if traits, ok := traitData.([]interface{}); ok {
					newTraits := make(map[string]interface{})
					for _, value := range traits {
						if typedValue, ok := value.([]interface{}); ok {
							if len(typedValue) != 3 {
								return fmt.Errorf("Parsed invalid trait (%v) in event: %s", value, data)
							}
							if traitType, ok := typedValue[1].(float64); ok {
								switch traitType {
								case 2:
									newTraits[typedValue[0].(string)] = typedValue[2].(float64)
								default:
									newTraits[typedValue[0].(string)] = typedValue[2].(string)
								}
							} else {
								return fmt.Errorf("Parsed invalid trait (%v) in event: %s", value, data)
							}
						} else {
							return fmt.Errorf("Parsed invalid trait (%v) in event: %s", value, data)
						}
					}
					newPayload["traits"] = newTraits
				}
			}
			for key, value := range typedPayload {
				if key != "traits" {
					newPayload[key] = value
				}
			}
			fmt.Printf("newPayload: %v\n", newPayload)
		}
		(*evt).parsed["payload"] = newPayload
	}

	return nil
}

//GeneratePrometheusAlert generates PrometheusAlert from the event data
func (evt *CeilometerEvent) GeneratePrometheusAlert(generatorURL string) PrometheusAlert {
	alert := PrometheusAlert{
		Labels:       make(map[string]string),
		Annotations:  make(map[string]string),
		GeneratorURL: generatorURL,
	}
	// set labels
	alert.Labels["alertname"] = evt.GetIndexName()
	surrogates := []AlertKeySurrogate{
		AlertKeySurrogate{"message_id", "messageId"},
		AlertKeySurrogate{"publisher_id", "instance"},
		AlertKeySurrogate{"event_type", "type"},
	}
	for _, renameCase := range surrogates {
		if value, ok := evt.parsed[renameCase.Parsed]; ok {
			alert.Labels[renameCase.Label] = value.(string)
		}
	}
	if value, ok := evt.parsed["priority"]; ok {
		if severity, ok := ceilometerAlertSeverity[value.(string)]; ok {
			alert.Labels["severity"] = severity
		} else {
			alert.Labels["severity"] = unknownSeverity
		}
	} else {
		alert.Labels["severity"] = unknownSeverity
	}
	if value, ok := evt.parsed["publisher_id"].(string); ok {
		alert.Labels["sourceName"] = strings.Join([]string{"ceilometer", value}, "@")
	}
	assimilateMap(evt.parsed["payload"].(map[string]interface{}), &alert.Annotations)
	// set timestamp
	if value, ok := evt.parsed["timestamp"].(string); ok {
		// ensure timestamp is in RFC3339
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, time.ANSIC, isoTimeLayout} {
			stamp, err := time.Parse(layout, value)
			if err == nil {
				alert.StartsAt = stamp.Format(time.RFC3339)
				break
			}
		}
	}
	// generate SG-relevant data
	alert.SetName()
	alert.SetSummary()
	alert.Labels["alertsource"] = "SmartGateway"
	return alert
}

//GeneratePrometheusAlertBody generates alert body for Prometheus Alert manager API
func (evt *CeilometerEvent) GeneratePrometheusAlertBody(generatorURL string) ([]byte, error) {
	return json.Marshal(evt.GeneratePrometheusAlert(generatorURL))
}
