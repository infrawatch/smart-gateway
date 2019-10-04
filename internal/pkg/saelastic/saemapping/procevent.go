package saemapping

import "time"

//ProceventMapping ...
const ProceventMapping = `
{
"procevent":{
   "mappings":{
      "event":{
         "properties":{
            "annotations":{
               "properties":{
                  "summary":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "ves":{
                     "properties":{
                        "domain":{
                           "type":"text",
                           "fields":{
                              "keyword":{
                                 "type":"keyword",
                                 "ignore_above":256
                              }
                           }
                        },
                        "eventId":{
                           "type":"long"
                        },
                        "eventName":{
                           "type":"text",
                           "fields":{
                              "keyword":{
                                 "type":"keyword",
                                 "ignore_above":256
                              }
                           }
                        },
                        "faultFields":{
                           "properties":{
                              "alarmCondition":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              },
                              "alarmInterfaceA":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              },
                              "eventSeverity":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              },
                              "eventSourceType":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              },
                              "faultFieldsVersion":{
                                 "type":"float"
                              },
                              "specificProblem":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              },
                              "vfStatus":{
                                 "type":"text",
                                 "fields":{
                                    "keyword":{
                                       "type":"keyword",
                                       "ignore_above":256
                                    }
                                 }
                              }
                           }
                        },
                        "lastEpochMicrosec":{
                           "type":"long"
                        },
                        "priority":{
                           "type":"text",
                           "fields":{
                              "keyword":{
                                 "type":"keyword",
                                 "ignore_above":256
                              }
                           }
                        },
                        "reportingEntityName":{
                           "type":"text",
                           "fields":{
                              "keyword":{
                                 "type":"keyword",
                                 "ignore_above":256
                              }
                           }
                        },
                        "sequence":{
                           "type":"long"
                        },
                        "sourceName":{
                           "type":"text",
                           "fields":{
                              "keyword":{
                                 "type":"keyword",
                                 "ignore_above":256
                              }
                           }
                        },
                        "startEpochMicrosec":{
                           "type":"long"
                        },
                        "version":{
                           "type":"float"
                        }
                     }
                  }
               }
            },
            "labels":{
               "properties":{
                  "alertname":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "instance":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "procevent":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "service":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "severity":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  },
                  "type":{
                     "type":"text",
                     "fields":{
                        "keyword":{
                           "type":"keyword",
                           "ignore_above":256
                        }
                     }
                  }
               }
            },
            "startsAt":{
               "type":"date"
            }
         }
      }
   }
}
}
}`

//ProcEvent ...
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
