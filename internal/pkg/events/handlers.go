package events

import (
	"encoding/json"
	"fmt"

	"github.com/infrawatch/smart-gateway/internal/pkg/events/incoming"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/infrawatch/smart-gateway/internal/pkg/saelastic"
)

// TODO: Implement this as pluggable system instead

//EventHandler provides interface for all possible handler types
type EventHandler interface {
	//Processes the event
	Handle(incoming.EventDataFormat, *saelastic.ElasticClient) (bool, error)
	//Relevant should return true if the handler is relevant for the givent event and so the handler should be used
	Relevant(incoming.EventDataFormat) bool
}

//EventHandlerManager holds all available handlers (and will be responsible
//in future for loading all handler plugins). The plugins will be organize
//per data source on which's events they could be applied
type EventHandlerManager struct {
	Handlers map[saconfig.DataSource][]EventHandler
}

//NewEventHandlerManager loads all even handler plugins stated in events configuration
func NewEventHandlerManager(config saconfig.EventConfiguration) (*EventHandlerManager, error) {
	manager := EventHandlerManager{}
	manager.Handlers = make(map[saconfig.DataSource][]EventHandler)
	for _, ds := range []saconfig.DataSource{saconfig.DataSourceCollectd, saconfig.DataSourceCeilometer, saconfig.DataSourceUniversal} {
		manager.Handlers[ds] = make([]EventHandler, 0)
	}

	for _, pluginPath := range config.HandlerPlugins {
		var ds saconfig.DataSource
		if ok := ds.SetFromString(pluginPath.DataSource); !ok {
			return &manager, fmt.Errorf("unknown datasource ''%s' for given event handler", pluginPath.DataSource)
		}
		manager.LoadHandlers(ds, pluginPath.Path)
	}

	//TODO: this just manually register the only handler we have now. Remove when the handler implementation will move out to plugin
	manager.Handlers[saconfig.DataSourceCollectd] = append(manager.Handlers[saconfig.DataSourceCollectd], ContainerHealthCheckHandler{"collectd_checks"})
	return &manager, nil
}

//LoadHandlers will load handler plugins in future
func (hand *EventHandlerManager) LoadHandlers(dataSource saconfig.DataSource, path string) error {

	return nil
}

//ContainerHealthCheckHandler serves as handler for events from collectd-sensubility's
//results of check-container-health.
type ContainerHealthCheckHandler struct {
	ElasticIndex string
}

type containerHealthCheckItem struct {
	Container string `json:"container"`
	Service   string `json:"service"`
	Status    string `json:"status"`
	Healthy   int    `json:"healthy"`
}

type list struct {
	Next *list
	Key  string
}

//recursivle search for output field based on path
func getOutputObject(path *list, data interface{}) (string, error) {
	var obj map[string]interface{}
	var ok bool

	if obj, ok = data.(map[string]interface{}); !ok {
		return "", fmt.Errorf("cannot search non-map objects")
	}
	if path.Next == nil {
		if output, ok := obj[path.Key].(string); ok {
			return output, nil
		}
		return "", fmt.Errorf("output should be of type 'string'")
	}
	if newData, ok := obj[path.Key]; ok {
		return getOutputObject(path.Next, newData)
	}
	return "", fmt.Errorf("input data does not contain path")
}

//Handle saves the event as separate document to ES in case the result output contains more than one item.
//Returns true if event processing should continue (eg. event should be saved to ES) or false if otherwise.
func (hand ContainerHealthCheckHandler) Handle(event incoming.EventDataFormat, elasticClient *saelastic.ElasticClient) (bool, error) {
	pathList := &list{
		Key: "annotations",
	}
	pathList.Next = &list{
		Key: "output",
	}

	if evt, ok := event.(*incoming.CollectdEvent); ok {
		rawData := evt.GetRawData()
		output, err := getOutputObject(pathList, rawData)
		if err != nil {
			return false, err
		}

		var outData []containerHealthCheckItem
		rawDataMap := rawData.(map[string]interface{})
		if err := json.Unmarshal([]byte(output), &outData); err == nil {
			for _, item := range outData {
				rawDataMap["annotations"].(map[string]interface{})["output"] = item
				if _, err := elasticClient.Create(hand.ElasticIndex, EVENTSINDEXTYPE, rawDataMap); err != nil {
					// saving the splitted output failed for some reason, so we will play safe
					// and try to process event outside of handler
					return true, err
				}
			}
		} else {
			// We most probably received single item output, so we just proceed and save the event
			if _, err := elasticClient.Create(hand.ElasticIndex, EVENTSINDEXTYPE, rawData); err != nil {
				return false, err
			}
		}
	}

	//record, err := elasticClient.Create(event.GetIndexName(), EVENTSINDEXTYPE, event.GetRawData())
	return false, nil
}

//Relevant returns true in case the event is suitable for processing with this handler, otherwise returns false.
func (hand ContainerHealthCheckHandler) Relevant(event incoming.EventDataFormat) bool {
	if evt, ok := event.(*incoming.CollectdEvent); ok {
		rawData := evt.GetRawData()
		if data, ok := rawData.(map[string]interface{}); ok {
			if rawLabels, ok := data["labels"]; ok {
				if labels, ok := rawLabels.(map[string]interface{}); ok {
					if check, ok := labels["check"]; ok {
						if checkName, ok := check.(string); ok && checkName == "check-container-health" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
