package saelastic


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
