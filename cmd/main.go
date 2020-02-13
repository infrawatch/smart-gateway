package main

import (
	"log"
	"os"
	"strings"

	"github.com/infrawatch/smart-gateway/internal/pkg/events"
	"github.com/infrawatch/smart-gateway/internal/pkg/metrics"
)

var metricType = "metrics"
var eventType = "events"

func main() {

	serviceType := getServiceType()
	if serviceType == metricType {
		metrics.StartMetrics()
	} else if serviceType == eventType {
		events.StartEvents()
	} else {
		log.Printf("Unknow command line argument 'servicetype' valid values are '%s' or '%s'.", metricType, eventType)
	}

}

//getServiceType ... checks for servicetype parameter to be either events or metrics.
func getServiceType() string {
	lenOfArgs := len(os.Args) - 1
	var servicetype = metricType
	for index, arg := range os.Args {
		pattern := "-servicetype="
		pposition := strings.Index(arg, pattern)
		if pposition > -1 {
			if pposition+len(pattern) < len(arg) {
				servicetype = arg[pposition+len(pattern):]
				return servicetype
			}
			log.Printf("Command line argument 'servicetype' %s is not valid.", arg)
			return ""
		}
		if arg == "-servicetype" || arg == "--servicetype" {
			if lenOfArgs >= index+1 {
				servicetype = os.Args[index+1]
				return servicetype
			}
			log.Printf("Command line argument 'servicetype' %s is not valid.", arg)
			return ""

		}
	}
	log.Printf("Command line argument 'servicetype' is not set, using default '%s' type.", servicetype)
	return servicetype
}
