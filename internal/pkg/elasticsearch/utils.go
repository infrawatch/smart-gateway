package saelastic

import (
	"encoding/json"
	"log"
	"strings"
)

//EventNotification ....
type EventNotification struct {
	Labels      func() interface{}
	Annotations func() interface{}
	StartsAt    string
}

//Sanitize ... Issue with json format in events.
func Sanitize(jsondata string) string {
	r := strings.NewReplacer("\\\"", "\"",
		"\"ves\":\"{", "\"ves\":{",
		"}\"}", "}}",
		"[", "",
		"]", "")
	result := r.Replace(jsondata)
	result1 := strings.Replace(result, "\\\"", "\"", -1)
	return result1

}

//GetIndexNameType ...
func GetIndexNameType(jsondata string) (string, IndexType, error) {
	//start := time.Now()

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
			switch val.(type) {
			case map[string]interface{}:
				if rec, ok := val.(map[string]interface{}); ok {
					if value, ok := rec["alertname"].(string); ok {
						index := strings.LastIndex(value, "_")
						if index > len("collectd_") {
							return value[0:index], EVENTINDEXTYPE, nil
						}
						return value, EVENTINDEXTYPE, nil
					}
					//else
					return string(GENERICINDEX), GENERICINDEXTYPE, nil
				} //else
				return string(GENERICINDEX), GENERICINDEXTYPE, nil
			}
		}
	default:
		return string(GENERICINDEX), GENERICINDEXTYPE, nil
	}
	return string(GENERICINDEX), GENERICINDEXTYPE, nil
}

// Not used
func typeSwitch(tst interface{}) (IndexName, IndexType, error) {
	switch v := tst.(type) {
	case map[string]interface{}:
		if val, ok := v["labels"]; ok {
			switch val.(type) {
			case map[string]interface{}:
				if rec, ok := val.(map[string]interface{}); ok {
					if _, ok := rec["connectivity"]; ok {
						return CONNECTIVITYINDEX, EVENTINDEXTYPE, nil
					} else if _, ok := rec["procevent"]; ok {
						return PROCEVENTINDEX, EVENTINDEXTYPE, nil
					} else if _, ok := rec["sysevent"]; ok {
						return SYSEVENTINDEX, EVENTINDEXTYPE, nil
					}
					//else
					return GENERICINDEX, GENERICINDEXTYPE, nil
				} //else
				return GENERICINDEX, GENERICINDEXTYPE, nil
			}
		}
	default:
		return GENERICINDEX, GENERICINDEXTYPE, nil
	}
	return GENERICINDEX, GENERICINDEXTYPE, nil
}
