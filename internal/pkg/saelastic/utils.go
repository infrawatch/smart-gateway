package saelastic

import (
	"encoding/json"
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

//SanitizeSummary ... summay is description and it adds double quote which breaks the json
func SanitizeSummary(text string) string {
	//look for summary and get the slice index
	reg := regexp.MustCompile(`summary":"`)
	indexes := reg.FindAllStringIndex(text, 1)
	if len(indexes) > 0 {
		//get slick index with first occurrence of comma with doublequote
		summaryreg := regexp.MustCompile("\",")
		summaryIndex := summaryreg.FindAllStringIndex(text[indexes[0][1]:len(text)], 1)
		// sanitize  everything within the summary
		summaryReplace := strings.NewReplacer("\"", "", "\\", "")
		// if summary has empty value then ignore example summary:"" else process
		if text[indexes[0][1]:indexes[0][1]+summaryIndex[0][0]] != "" {
			//sanitize result from end of summary key slice to value slice
			sanresult := summaryReplace.Replace(text[indexes[0][1] : indexes[0][1]+summaryIndex[0][0]])
			//make a new slice and allocate memory to the length of the text( this needs to re calculated ,else end up with extra spaces , )
			result := make([]string, len(text)+1)
			//get the first part
			result[0] = text[0:indexes[0][1]]
			// add the satized part
			result[1] = sanresult
			// suffice the remaining part
			result[2] = text[indexes[0][1]+summaryIndex[0][0]:]
			//convert as string
			return strings.Join(result, "")
		}

	}

	return text

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
	result1 = SanitizeSummary(result1)

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
