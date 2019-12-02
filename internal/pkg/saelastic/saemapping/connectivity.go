package saemapping

import "time"

//REMOVE: Whole file. Not used anywhere

//ConnectivityMapping ...
const ConnectivityMapping = `
{
   "mappings": {
      "event": {
         "properties": {
            "labels": {
               "type": "object",
               "properties": {
                  "alertname": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                           "ignore_above":	256
                          }
                        }
                  },
                  "instance": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword"
                          }
                        }
                  },
                  "connectivity": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                          "ignore_above":	256
                          }
                        }
                  },
                  "type": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                           "ignore_above":	256
                          }
                        }
                  },
                  "severity": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                           "ignore_above":	256
                          }
                        }
                  },
                  "service": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                           "ignore_above":	256
                          }
                        }
                  }
               }
            },
            "annotations": {
               "type": "object",
               "properties": {
                  "summary": {
                     "type": "text",
                     "fields": {
                          "keyword": {
                          "type": "keyword",
                           "ignore_above":	256
                          }
                        }
                  },
                  "ves": {
                     "type": "object",
                     "properties": {
                        "domain": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "stateChange": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "eventId": {
                           "type": "long"
                        },
                        "eventName": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "lastEpochMicrosec": {
                           "type": "long"
                        },
                        "priority": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "reportingEntityName": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "sequence": {
                           "type": "long"
                        },
                        "sourceName": {
                           "type": "text",
                           "fields": {
                                "keyword": {
                                "type": "keyword",
                                 "ignore_above":	256
                                }
                              }
                        },
                        "startEpochMicrosec": {
                           "type": "long"

                        },
                        "version": {
                           "type": "float"
                        },
                        "stateChangeFields": {
                           "type": "object",
                           "properties": {
                              "newState": {
                                 "type": "text",
                                  "index": true
                              },
                              "oldState": {
                                  "type": "text",
                                  "fields": {
                                       "keyword": {
                                       "type": "keyword",
                                        "ignore_above":	256
                                       }
                                     }
                              },
                              "stateChangeFieldsVersion": {
                                 "type": "float"
                              }
                           }
                        }
                     }
                  }
               }
            },
            "startsAt": {
               "type": "date"
            }
         }
       }
      }
}
`

//Connectivity ... generated using https://mholt.github.io/json-to-go
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
