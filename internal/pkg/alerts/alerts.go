package alerts

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
)

//Alerts  ...
type Alerts struct {
	//Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL"`
}

var alertSeverity = map[string]string{
	"OKAY":    "info",
	"WARNING": "warning",
	"FAILURE": "critical",
}
var alertStatus = map[string]string{
	"OKAY":    "resolved",
	"WARNING": "firing",
	"FAILURE": "firing",
}

//SetSummary ...
func (a *Alerts) SetSummary() {
	values := make([]string, 0, 4)
	if val, ok := a.Labels["summary"]; ok {
		a.Annotations["summary"] = val
	}
	if a.Annotations["summary"] == "" {
		if sourceVal, sourceOk := a.Labels["sourceName"]; sourceOk {
			values = append(values, sourceVal)
		}
		if typeVal, typeOk := a.Labels["type"]; typeOk {
			values = append(values, typeVal)
		}
		if eventNameVal, eventNameOk := a.Labels["eventName"]; eventNameOk {
			values = append(values, eventNameVal)
		}
	}

	//log.Printf("alert summary %s", values)
	a.Annotations["summary"] = strings.Join(values, " ")
}

//SetName ... set unique name for alerts
func (a *Alerts) SetName() {

	values := make([]string, 0, len(a.Labels)-1)
	desc := make([]string, 0, len(a.Labels)-1)
	keys := make([]string, 0, len(a.Labels))
	for k := range a.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if k != "severity" {
			values = append(values, a.Labels[k])
			desc = append(desc, a.Labels[k])
		} else {
			if val, ok := alertSeverity[a.Labels[k]]; ok {
				desc = append(desc, val)
			} else {
				desc = append(desc, a.Labels[k])
			}
		}

	}
	a.Labels["name"] = strings.Join(values, "_")
	a.Annotations["description"] = strings.Join(desc, " ")
}

//Parse ...parses alerts to validate for schema
func (a *Alerts) Parse(eventJSON []byte, generatorURL string) {
	//form collectd An associated severity, which can be one of OKAY, WARNING, and FAILURE.
	var dat []map[string]interface{}
	a.GeneratorURL = generatorURL
	if err := json.Unmarshal(eventJSON, &dat); err != nil {
		log.Println("Error parsing events for alerts.")
		log.Panic(err)
	}
	a.Annotations = make(map[string]string)
	a.Labels = make(map[string]string)
	labels := dat[0]["labels"].(map[string]interface{})
	annotaions := dat[0]["annotations"].(map[string]interface{})
	for k, v := range labels {
		if k == "severity" {
			if valSeverity, okSeverity := alertSeverity[v.(string)]; okSeverity {
				a.Labels[k] = valSeverity
			} else {
				a.Labels[k] = v.(string)
			}
			/*if valStatus, okStatus := alertStatus[v.(string)]; okStatus {
				a.Status = valStatus
			} else {
				a.Status = "firing"
			}*/
		} else {
			a.Labels[k] = v.(string)
		}
	}
	a.SetName()
	a.Labels["alertsource"] = "SMARTAGENT"
	a.Labels["service"] = "SmartGateway"
	a.Labels["location"] = labels["instance"].(string)

	for key, val := range annotaions {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			a.parseMap(val.(map[string]interface{}))
		case []interface{}:
			//do nothing
		default:
			a.Labels[key] = concreteVal.(string)

		}
	}
	a.SetSummary()

}

func (a *Alerts) parseMap(aMap map[string]interface{}) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			fmt.Println(key)
			a.parseMap(val.(map[string]interface{}))
		case []interface{}:
			//donothing now
		default:
			a.Labels[key] = concreteVal.(string)
		}
	}
}
