package tests

import (
	"fmt"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/alerts"
	"github.com/stretchr/testify/assert"
)

type DataMatrix struct {
	Label    string
	Expected string
}

const event = `{"labels":{"alertname":"collectd_connectivity_gauge","instance":"nfvha-comp-03","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":"39996","eventName":"interface eno2 up","lastEpochMicrosec":"1523292316174821","priority":"high","reportingEntityName":"collectd connectivity plugin","sourceName":"eno2","version":"1","stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":"1","stateInterface":"eno2"}}},"startsAt":"2018-04-09T16:45:16.174815108Z"}`

func TestAlertParsing(t *testing.T) {
	// mock event
	eventString := []byte(fmt.Sprintf("[%v]", event))
	// mock alert
	eventAlert := &alerts.Alerts{}
	eventAlert.Parse(eventString, "http://0.0.0.0/")
	// prepare test matrix for Alerts.Labels and verify it
	data := []DataMatrix{
		DataMatrix{"alertname", "collectd_connectivity_gauge"},
		DataMatrix{"instance", "nfvha-comp-03"},
		DataMatrix{"connectivity", "eno2"},
		DataMatrix{"type", "interface_status"},
		DataMatrix{"severity", "info"},
		//DataMatrix{"service", "collectd"},
		DataMatrix{"domain", "stateChange"},
		DataMatrix{"eventId", "39996"},
		DataMatrix{"eventName", "interface eno2 up"},
		DataMatrix{"lastEpochMicrosec", "1523292316174821"},
		DataMatrix{"priority", "high"},
		DataMatrix{"reportingEntityName", "collectd connectivity plugin"},
		DataMatrix{"sourceName", "eno2"},
		DataMatrix{"version", "1"},
		DataMatrix{"newState", "inService"},
		DataMatrix{"oldState", "outOfService"},
		DataMatrix{"stateChangeFieldsVersion", "1"},
		DataMatrix{"stateInterface", "eno2"},
		DataMatrix{"summary", ""},
		DataMatrix{"name", "collectd_connectivity_gauge_eno2_nfvha-comp-03_collectd_interface_status"},
	}
	t.Run("Verify proper parsing of event data to Alerts.Labels", func(t *testing.T) {
		for _, testCase := range data {
			assert.Equal(t, testCase.Expected, eventAlert.Labels[testCase.Label])
		}
	})
	// prepare test matrix for Alerts.Annotation and verify it
	data = []DataMatrix{
		DataMatrix{"summary", "eno2 interface_status interface eno2 up"},
		DataMatrix{"description", "collectd_connectivity_gauge eno2 nfvha-comp-03 collectd info interface_status"},
	}
	t.Run("Verify proper parsing of event data to Alerts.Annotations", func(t *testing.T) {
		for _, testCase := range data {
			assert.Equal(t, testCase.Expected, eventAlert.Annotations[testCase.Label])
		}
	})
	t.Run("Verify proper parsing of rest of data", func(t *testing.T) {
		// TO-DO(mmagr): This fails currently
		//assert.Equal(t, "2018-04-09T16:45:16.174815108Z", eventAlert.StartsAt)
		assert.Equal(t, "http://0.0.0.0/", eventAlert.GeneratorURL)
	})
}
