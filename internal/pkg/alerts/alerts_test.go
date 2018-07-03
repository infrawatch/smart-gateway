package alerts

import (
	"encoding/json"
	"log"
	"testing"
)

func TestAlertParsing(t *testing.T) {
	var event = `[{"labels":{"alertname":"collectd_connectivity_gauge","instance":"nfvha-comp-03","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":"39996","eventName":"interface eno2 up","lastEpochMicrosec":"1523292316174821","priority":"high","reportingEntityName":"collectd connectivity plugin","sourceName":"eno2","version":"1","stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":"1","stateInterface":"eno2"}}},"startsAt":"2018-04-09T16:45:16.174815108Z"}]`
	var eventString = []byte(event)
	var eventAlert = &Alerts{}
	eventAlert.Parse(eventString, "http://0.0.0.0/")
	jsonString, err := json.Marshal(eventAlert)
	if err != nil {
		panic(err)
	}
	var s = []string{string(jsonString)}
	log.Printf("%#v", eventAlert)
	log.Printf("%s", s)
	log.Printf("%s", err)

}
