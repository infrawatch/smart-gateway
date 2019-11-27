package saelastic

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

//EventNotification ....
type EventNotification struct {
	Labels      func() interface{}
	Annotations func() interface{}
	StartsAt    string
}

var rexForNestedQuote = regexp.MustCompile(`\\\"`)
var rexForVes = regexp.MustCompile(`"ves":"{(.*)}"`)

//Sanitize ... Issue with json format in events.
func Sanitize(jsondata string) string {
	// if value for key "ves" is string, we convert it to json
	vesCleaned := jsondata
	sub := rexForVes.FindStringSubmatch(jsondata)
	if len(sub) == 2 {
		nested := rexForNestedQuote.ReplaceAllString(sub[1], `"`)
		vesCleaned = rexForVes.ReplaceAllString(jsondata, fmt.Sprintf(`"ves":{%s}`, nested))
	}

	almostFinal := strings.TrimLeft(vesCleaned, "[")
	return strings.TrimRight(almostFinal, "]")
}

//GetIndexNameType ...
func GetIndexNameType(jsondata string) (string, IndexType, error) {
	var f []interface{}
	err := json.Unmarshal([]byte(jsondata), &f)
	if err != nil {
		log.Fatal(err)
		return string(GENERICINDEX), GENERICINDEXTYPE, err
	}
	//elapsed := time.Since(start)
	index, indextype, error := typeSwitchAlertname(f[0])
	//	log.Printf("getIndexNameType took %s", elapsed)
	return index, indextype, error
}

func typeSwitchAlertname(tst interface{}) (string, IndexType, error) {
	switch v := tst.(type) {
	case map[string]interface{}:
		if val, ok := v["labels"]; ok {
			switch rec := val.(type) {
			case map[string]interface{}:
				if value, ok := rec["alertname"].(string); ok {
					index := strings.LastIndex(value, "_")
					if index > len("collectd_") {
						return value[0:index], EVENTINDEXTYPE, nil
					}
					return value, EVENTINDEXTYPE, nil
				}
				//else
				return string(GENERICINDEX), GENERICINDEXTYPE, nil
			}
		}
	default:
		return string(GENERICINDEX), GENERICINDEXTYPE, nil
	}
	return string(GENERICINDEX), GENERICINDEXTYPE, nil
}

//lint:ignore U1000 we might be able to delete this, but we should dig into the history of what it's for
func typeSwitch(tst interface{}) (IndexName, IndexType, error) {
	switch v := tst.(type) {
	case map[string]interface{}:
		if val, ok := v["labels"]; ok {
			switch rec := val.(type) {
			case map[string]interface{}:
				if _, ok := rec["connectivity"]; ok {
					return CONNECTIVITYINDEX, EVENTINDEXTYPE, nil
				} else if _, ok := rec["procevent"]; ok {
					return PROCEVENTINDEX, EVENTINDEXTYPE, nil
				} else if _, ok := rec["sysevent"]; ok {
					return SYSEVENTINDEX, EVENTINDEXTYPE, nil
				}
				//else
				return GENERICINDEX, GENERICINDEXTYPE, nil
			}
		}
	default:
		return GENERICINDEX, GENERICINDEXTYPE, nil
	}
	return GENERICINDEX, GENERICINDEXTYPE, nil
}
