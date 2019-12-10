package tests

import (
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/events/incoming"
	"github.com/stretchr/testify/assert"
)

const (
	// alert testing event
	eventForAlert = `{"labels":{"alertname":"collectd_connectivity_gauge","instance":"nfvha-comp-03","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":"39996","eventName":"interface eno2 up","lastEpochMicrosec":"1523292316174821","priority":"high","reportingEntityName":"collectd connectivity plugin","sourceName":"eno2","version":"1","slicetest":["item1","item2","item3"],"stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":"1","stateInterface":"eno2"}}},"startsAt":"2018-04-09T16:45:16.174815108Z"}`
	// collectd messages
	connectivityEventData = "[{\"labels\":{\"alertname\":\"collectd_connectivity_gauge\",\"instance\":\"d60b3c68f23e\",\"connectivity\":\"eno2\",\"type\":\"interface_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"stateChange\\\",\\\"eventId\\\":2,\\\"eventName\\\":\\\"interface eno2 up\\\",\\\"lastEpochMicrosec\\\":1518790014024924,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd connectivity plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"eno2\\\",\\\"startEpochMicrosec\\\":1518790009881440,\\\"version\\\":1.0,\\\"stateChangeFields\\\":{\\\"newState\\\":\\\"outOfService\\\",\\\"oldState\\\":\\\"inService\\\",\\\"stateChangeFieldsVersion\\\":1.0,\\\"stateInterface\\\":\\\"eno2\\\"}}\"},\"startsAt\":\"2018-02-16T14:06:54.024856417Z\"}]"
	procEventData1        = "[{\"labels\":{\"alertname\":\"collectd_procevent_gauge\",\"instance\":\"d60b3c68f23e\",\"procevent\":\"bla.py\",\"type\":\"process_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"fault\\\",\\\"eventId\\\":3,\\\"eventName\\\":\\\"process bla.py (8537) down\\\",\\\"lastEpochMicrosec\\\":1518791119579620,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd procevent plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"bla.py\\\",\\\"startEpochMicrosec\\\":1518791111336973,\\\"version\\\":1.0,\\\"faultFields\\\":{\\\"alarmCondition\\\":\\\"process bla.py (8537) state change\\\",\\\"alarmInterfaceA\\\":\\\"bla.py\\\",\\\"eventSeverity\\\":\\\"CRITICAL\\\",\\\"eventSourceType\\\":\\\"process\\\",\\\"faultFieldsVersion\\\":1.0,\\\"specificProblem\\\":\\\"process bla.py (8537) down\\\",\\\"vfStatus\\\":\\\"Ready to terminate\\\"}}\"},\"startsAt\":\"2018-02-16T14:25:19.579573212Z\"}]"
	procEventData2        = `[{"labels":{"alertname":"collectd_interface_if_octets","instance":"localhost.localdomain","interface":"lo","severity":"FAILURE","service":"collectd"},"annotations":{"summary":"Host localhost.localdomain, plugin interface (instance lo) type if_octets: Data source \"rx\" is currently 43596.224329. That is above the failure threshold of 0.000000.","DataSource":"rx","CurrentValue":"43596.2243286703","WarningMin":"nan","WarningMax":"nan","FailureMin":"nan","FailureMax":"0"},"startsAt":"2019-09-18T21:11:19.281603240Z"}]`
	ovsEventData          = `[{"labels":{"alertname":"collectd_ovs_events_gauge","instance":"nfvha-comp-03","ovs_events":"br0","type":"link_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"link state of \"br0\" interface has been changed to \"UP\"","uuid":"c52f2aca-3cb1-48e3-bba3-100b54303d84"},"startsAt":"2018-02-22T20:12:19.547955618Z"}]`
	// ceilometer messages
	// generic messages
)

var (
	// collectd events
	connectivityEvent = map[string]interface{}{
		"labels": map[string]interface{}{
			"alertname":    "collectd_connectivity_gauge",
			"instance":     "d60b3c68f23e",
			"connectivity": "eno2",
			"type":         "interface_status",
			"severity":     "FAILURE",
			"service":      "collectd",
		},
		"annotations": map[string]interface{}{
			"summary": "",
			"ves": map[string]interface{}{
				"domain":              "stateChange",
				"eventId":             2.0,
				"eventName":           "interface eno2 up",
				"lastEpochMicrosec":   float64(1518790014024924),
				"priority":            "high",
				"reportingEntityName": "collectd connectivity plugin",
				"sequence":            0.0,
				"sourceName":          "eno2",
				"startEpochMicrosec":  float64(1518790009881440),
				"version":             1.0,
				"stateChangeFields": map[string]interface{}{
					"newState":                 "outOfService",
					"oldState":                 "inService",
					"stateChangeFieldsVersion": 1.0,
					"stateInterface":           "eno2",
				},
			},
		},
		"startsAt": "2018-02-16T14:06:54.024856417Z",
	}
	procEvent1 = map[string]interface{}{
		"labels": map[string]interface{}{
			"alertname": "collectd_procevent_gauge",
			"instance":  "d60b3c68f23e",
			"procevent": "bla.py",
			"type":      "process_status",
			"severity":  "FAILURE",
			"service":   "collectd",
		},
		"annotations": map[string]interface{}{
			"summary": "",
			"ves": map[string]interface{}{
				"domain":              "fault",
				"eventId":             3.0,
				"eventName":           "process bla.py (8537) down",
				"lastEpochMicrosec":   float64(1518791119579620),
				"priority":            "high",
				"reportingEntityName": "collectd procevent plugin",
				"sequence":            0.0,
				"sourceName":          "bla.py",
				"startEpochMicrosec":  float64(1518791111336973),
				"version":             1.0,
				"faultFields": map[string]interface{}{
					"alarmCondition":     "process bla.py (8537) state change",
					"alarmInterfaceA":    "bla.py",
					"eventSeverity":      "CRITICAL",
					"eventSourceType":    "process",
					"faultFieldsVersion": 1.0,
					"specificProblem":    "process bla.py (8537) down",
					"vfStatus":           "Ready to terminate",
				},
			},
		},
		"startsAt": "2018-02-16T14:25:19.579573212Z",
	}
	procEvent2 = map[string]interface{}{
		"labels": map[string]interface{}{
			"alertname": "collectd_interface_if_octets",
			"instance":  "localhost.localdomain",
			"interface": "lo",
			"severity":  "FAILURE",
			"service":   "collectd",
		},
		"annotations": map[string]interface{}{
			"summary":      "Host localhost.localdomain, plugin interface (instance lo) type if_octets: Data source \"rx\" is currently 43596.224329. That is above the failure threshold of 0.000000.",
			"DataSource":   "rx",
			"CurrentValue": "43596.2243286703",
			"WarningMin":   "nan",
			"WarningMax":   "nan",
			"FailureMin":   "nan",
			"FailureMax":   "0",
		},
		"startsAt": "2019-09-18T21:11:19.281603240Z",
	}
	ovsEvent = map[string]interface{}{
		"labels": map[string]interface{}{
			"alertname":  "collectd_ovs_events_gauge",
			"instance":   "nfvha-comp-03",
			"ovs_events": "br0",
			"type":       "link_status",
			"severity":   "OKAY",
			"service":    "collectd",
		},
		"annotations": map[string]interface{}{
			"summary": "link state of \"br0\" interface has been changed to \"UP\"",
			"uuid":    "c52f2aca-3cb1-48e3-bba3-100b54303d84",
		},
		"startsAt": "2018-02-22T20:12:19.547955618Z",
	}
	// ceilometer events
	// generic events
)

type EventDataParsingTestMatrix struct {
	Dirty     string
	Sanitized map[string]interface{}
	IndexName string
}

func TestEventDataParsing(t *testing.T) {
	collectdMatrix := []EventDataParsingTestMatrix{
		{procEventData1, procEvent1, "collectd_procevent"},
		{procEventData2, procEvent2, "collectd_interface_if"},
		{ovsEventData, ovsEvent, "collectd_ovs_events"},
		{connectivityEventData, connectivityEvent, "collectd_connectivity"},
	}
	for _, testCase := range collectdMatrix {
		evt := incoming.CollectdEvent{}
		evt.ParseEvent(testCase.Dirty)
		assert.Equal(t, testCase.Sanitized, evt.GetRawData())
		assert.Equal(t, testCase.IndexName, evt.GetIndexName())
	}
}

type EventAlertDataMatrix struct {
	Label    string
	Expected string
}

func TestGenerateAlert(t *testing.T) {
	// mock event
	evt := incoming.CollectdEvent{}
	evt.ParseEvent(eventForAlert)
	// mock alert
	eventAlert := evt.GeneratePrometheusAlert("https://this/is/test")
	// prepare test matrix for Alerts.Labels and verify it
	data := []EventAlertDataMatrix{
		{"alertname", "collectd_connectivity_gauge"},
		{"instance", "nfvha-comp-03"},
		{"connectivity", "eno2"},
		{"type", "interface_status"},
		{"severity", "info"},
		//AlertDataMatrix{"service", "collectd"},
		{"domain", "stateChange"},
		{"eventId", "39996"},
		{"eventName", "interface eno2 up"},
		{"lastEpochMicrosec", "1523292316174821"},
		{"priority", "high"},
		{"reportingEntityName", "collectd connectivity plugin"},
		{"sourceName", "eno2"},
		{"version", "1"},
		{"newState", "inService"},
		{"oldState", "outOfService"},
		{"stateChangeFieldsVersion", "1"},
		{"stateInterface", "eno2"},
		{"summary", ""},
		{"name", "collectd_connectivity_gauge_eno2_nfvha-comp-03_collectd_interface_status"},
		{"slicetest", "item1,item2,item3"},
	}
	t.Run("Verify proper parsing of event data to Labels", func(t *testing.T) {
		for _, testCase := range data {
			assert.Equalf(t, testCase.Expected, eventAlert.Labels[testCase.Label], "Unexpected label for %s", testCase.Label)
		}
	})
	// prepare test matrix for Alerts.Annotation and verify it
	data = []EventAlertDataMatrix{
		{"summary", "eno2 interface_status interface eno2 up"},
		{"description", "collectd_connectivity_gauge eno2 nfvha-comp-03 collectd info interface_status"},
	}
	t.Run("Verify proper parsing of event data to Annotations", func(t *testing.T) {
		for _, testCase := range data {
			assert.Equalf(t, testCase.Expected, eventAlert.Annotations[testCase.Label], "Unexpected annotation for %s", testCase.Label)
		}
	})
	t.Run("Verify proper parsing of rest of data", func(t *testing.T) {
		assert.Equal(t, "2018-04-09T16:45:16Z", eventAlert.StartsAt)
		assert.Equal(t, "https://this/is/test", eventAlert.GeneratorURL)
	})
}
