package incoming

import (
	"sort"
	"strings"
)

type EventDataFormat interface {
	//GetIndexName returns Elasticsearch index to which this event is or should be saved.
	GetIndexName() string
	//GetRawData returns sanitized and umarshalled event data.
	GetRawData() interface{}
	//GetSanitized returns sanitized event data
	GetSanitized() string
	//ParseEvent sanitizes and unmarshals received event data.
	ParseEvent(string) error
	//GeneratePrometheusAlertBody generates alert body for Prometheus Alert manager API
	GeneratePrometheusAlertBody(string) ([]byte, error)
}

type PrometheusAlert struct {
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorUrl string            `json:"generatorURL"`
}

//SetName generates unique name and description for the alert and creates new key/value pair for it in Labels
func (self *PrometheusAlert) SetName() {
	if _, ok := self.Labels["name"]; !ok {
		keys := make([]string, 0, len(self.Labels))
		for k := range self.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		values := make([]string, 0, len(self.Labels)-1)
		desc := make([]string, 0, len(self.Labels))
		for _, k := range keys {
			if k != "severity" {
				values = append(values, self.Labels[k])
			}
			desc = append(desc, self.Labels[k])
		}
		self.Labels["name"] = strings.Join(values, "_")
		self.Annotations["description"] = strings.Join(desc, " ")
	}
}

//SetSummary generates summary annotation in case it is empty
func (self *PrometheusAlert) SetSummary() {
	generate := false
	if _, ok := self.Annotations["summary"]; ok {
		if self.Annotations["summary"] == "" {
			generate = true
		}
	} else {
		generate = true
	}

	if generate {
		if val, ok := self.Labels["summary"]; ok && self.Labels["summary"] != "" {
			self.Annotations["summary"] = val
		} else {
			values := make([]string, 0, 3)
			for _, key := range []string{"sourceName", "type", "eventName"} {
				if val, ok := self.Labels[key]; ok {
					values = append(values, val)
				}
			}
			self.Annotations["summary"] = strings.Join(values, " ")
		}
	}
}
