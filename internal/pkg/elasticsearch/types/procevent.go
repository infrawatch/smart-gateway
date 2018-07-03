package saelastic
import (
  "time"
)

//ProcEvent ....
type ProcEvent []struct {
Labels struct {
  Alertname string `json:"alertname"`
  Instance  string `json:"instance"`
  Procevent string `json:"procevent"`
  Type      string `json:"type"`
  Severity  string `json:"severity"`
  Service   string `json:"service"`
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
    FaultFields         struct {
      AlarmCondition     string  `json:"alarmCondition"`
      AlarmInterfaceA    string  `json:"alarmInterfaceA"`
      EventSeverity      string  `json:"eventSeverity"`
      EventSourceType    string  `json:"eventSourceType"`
      FaultFieldsVersion float64 `json:"faultFieldsVersion"`
      SpecificProblem    string  `json:"specificProblem"`
      VfStatus           string  `json:"vfStatus"`
    } `json:"faultFields"`
  } `json:"ves"`
} `json:"annotations"`
StartsAt time.Time `json:"startsAt"`
}
