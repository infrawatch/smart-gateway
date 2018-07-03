package saelastic
import (
  "time"
)
/** generated using https://mholt.github.io/json-to-go*/

//Connectivity .....
type Connectivity []struct {
	Labels struct {
		Alertname    string `json:"alertname"`
		Instance     string `json:"instance"`
		Connectivity string `json:"connectivity"`
		Type         string `json:"type"`
		Severity     string `json:"severity"`
		Service      string `json:"service"`
	} `json:"labels"`
	Annotations struct {
		Summary string `json:"summary"`
		Ves     struct {
			Domain              string  `json:"domain"`
			EventID             int     `json:"eventId"`
			EventName           string  `json:"eventName"`
			LastEpochMicrosec   int64   `json:"lastEpochMicrosec"`
			Priority            string  `json:"priority"`
			ReportingEntityName string  `json:"reportingEntityName"`
			Sequence            int     `json:"sequence"`
			SourceName          string  `json:"sourceName"`
			StartEpochMicrosec  int64   `json:"startEpochMicrosec"`
			Version             float64 `json:"version"`
			StateChangeFields   struct {
				NewState                 string  `json:"newState"`
				OldState                 string  `json:"oldState"`
				StateChangeFieldsVersion float64 `json:"stateChangeFieldsVersion"`
				StateInterface           string  `json:"stateInterface"`
			} `json:"stateChangeFields"`
		} `json:"ves"`
	} `json:"annotations"`
	StartsAt time.Time `json:"startsAt"`
}
