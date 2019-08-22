package tests

import (
	"fmt"
	"log"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic/saemapping"
)

//COLLECTD
const (
	CONNECTIVITYINDEXTEST = "collectd_connectivity_test"
	PROCEVENTINDEXTEST    = "collectd_procevent_test"
	SYSEVENTINDEXTEST     = "collectd_syslogs_test"
	GENERICINDEXTEST      = "collectd_generic_test"
	connectivitydata      = `[{"labels":{"alertname":"collectd_connectivity_gauge","instance":"d60b3c68f23e","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":2,"eventName":"interface eno2 up","lastEpochMicrosec":1518188764024922,"priority":"high","reportingEntityName":"collectd connectivity plugin","sequence":0,"sourceName":"eno2","startEpochMicrosec":1518188755700851,"version":1.0,"stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":1.0,"stateInterface":"eno2"}}},"startsAt":"2018-02-09T15:06:04.024859063Z"}]`
	connectivityDirty     = "[{\"labels\":{\"alertname\":\"collectd_connectivity_gauge\",\"instance\":\"d60b3c68f23e\",\"connectivity\":\"eno2\",\"type\":\"interface_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"stateChange\\\",\\\"eventId\\\":11,\\\"eventName\\\":\\\"interface eno2 down\\\",\\\"lastEpochMicrosec\\\":1518790014024924,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd connectivity plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"eno2\\\",\\\"startEpochMicrosec\\\":1518790009881440,\\\"version\\\":1.0,\\\"stateChangeFields\\\":{\\\"newState\\\":\\\"outOfService\\\",\\\"oldState\\\":\\\"inService\\\",\\\"stateChangeFieldsVersion\\\":1.0,\\\"stateInterface\\\":\\\"eno2\\\"}}\"},\"startsAt\":\"2018-02-16T14:06:54.024856417Z\"}]"
	procEventData         = `[{"labels":{"alertname":"collectd_procevent_gauge","instance":"d60b3c68f23e","procevent":"bla.py","type":"process_status","severity":"FAILURE","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"fault","eventId":3,"eventName":"process bla.py (8537) down","lastEpochMicrosec":1518791119579620,"priority":"high","reportingEntityName":"collectd procevent plugin","sequence":0,"sourceName":"bla.py","startEpochMicrosec":1518791111336973,"version":1.0,"faultFields":{"alarmCondition":"process bla.py (8537) state change","alarmInterfaceA":"bla.py","eventSeverity":"CRITICAL","eventSourceType":"process","faultFieldsVersion":1.0,"specificProblem":"process bla.py (8537) down","vfStatus":"Ready to terminate"}}},"startsAt":"2018-02-16T14:25:19.579573212Z"}]`
	procEventDirty        = "[{\"labels\":{\"alertname\":\"collectd_procevent_gauge\",\"instance\":\"d60b3c68f23e\",\"procevent\":\"bla.py\",\"type\":\"process_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"fault\\\",\\\"eventId\\\":3,\\\"eventName\\\":\\\"process bla.py (8537) down\\\",\\\"lastEpochMicrosec\\\":1518791119579620,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd procevent plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"bla.py\\\",\\\"startEpochMicrosec\\\":1518791111336973,\\\"version\\\":1.0,\\\"faultFields\\\":{\\\"alarmCondition\\\":\\\"process bla.py (8537) state change\\\",\\\"alarmInterfaceA\\\":\\\"bla.py\\\",\\\"eventSeverity\\\":\\\"CRITICAL\\\",\\\"eventSourceType\\\":\\\"process\\\",\\\"faultFieldsVersion\\\":1.0,\\\"specificProblem\\\":\\\"process bla.py (8537) down\\\",\\\"vfStatus\\\":\\\"Ready to terminate\\\"}}\"},\"startsAt\":\"2018-02-16T14:25:19.579573212Z\"}]"
	ovsEvent              = `[{"labels":{"alertname":"collectd_ovs_events_gauge","instance":"nfvha-comp-03","ovs_events":"br0","type":"link_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"link state of \"br0\" interface has been changed to \"UP\"","uuid":"c52f2aca-3cb1-48e3-bba3-100b54303d84"},"startsAt":"2018-02-22T20:12:19.547955618Z"}]`
	elastichost           = "http://127.0.0.1:9200"
)

func TestMain(t *testing.T) {
	config := saconfig.EventConfiguration{
		Debug:          false,
		ElasticHostURL: elastichost,
		UseTls:         false,
		TlsClientCert:  "",
		TlsClientKey:   "",
		TlsCaCert:      "",
	}

	client, err := saelastic.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to connect to elastic search: %s", err)
	} else {
		defer func() {
			client.DeleteIndex(string(CONNECTIVITYINDEXTEST))
			client.DeleteIndex(string(PROCEVENTINDEXTEST))
			client.DeleteIndex(string(SYSEVENTINDEXTEST))
			client.DeleteIndex(string(GENERICINDEXTEST))
		}()
	}

	t.Run("Test create and delete", func(t *testing.T) {
		indexName, _, err := saelastic.GetIndexNameType(connectivitydata)
		if err != nil {
			t.Errorf("Failed to get indexname and type%s", err)
			return
		}

		testIndexname := fmt.Sprintf("%s_%s", indexName, "test")
		client.DeleteIndex(testIndexname)
		client.CreateIndex(testIndexname, saemapping.ConnectivityMapping)
		exists, err := client.IndexExists(string(testIndexname)).Do(client.GetContext())
		if exists == false || err != nil {
			t.Errorf("Failed to create index %s", err)
		}
		err = client.DeleteIndex(testIndexname)
		if err != nil {
			t.Errorf("Failed to Delete index %s", err)
		}
	})

	t.Run("Test connectivity data create", func(t *testing.T) {
		indexName, IndexType, err := saelastic.GetIndexNameType(connectivitydata)
		if err != nil {
			t.Errorf("Failed to get indexname and type%s", err)
			return
		}
		testIndexname := fmt.Sprintf("%s_%s", indexName, "test")
		err = client.DeleteIndex(testIndexname)

		client.CreateIndex(testIndexname, saemapping.ConnectivityMapping)
		exists, err := client.IndexExists(string(testIndexname)).Do(client.GetContext())
		if exists == false || err != nil {
			t.Errorf("Failed to create index %s", err)
		}

		id, err := client.Create(testIndexname, IndexType, connectivitydata)
		if err != nil {
			t.Errorf("Failed to create data %s\n", err.Error())
		} else {
			log.Printf("document id  %#v\n", id)
		}
		result, err := client.Get(testIndexname, IndexType, id)
		if err != nil {
			t.Errorf("Failed to get data %s", err)
		} else {
			log.Printf("Data %#v", result)
		}
		deleteErr := client.Delete(testIndexname, IndexType, id)
		if deleteErr != nil {
			t.Errorf("Failed to delete data %s", deleteErr)
		}

		err = client.DeleteIndex(testIndexname)
		if err != nil {
			t.Errorf("Failed to Delete index %s", err)
		}
	})
}

/*func TestIndexCheckConnectivity(t *testing.T) {
	indexName, indexType, err := saelastic.GetIndexNameType(connectivitydata)
	if err != nil {
		t.Errorf("Failed to get indexname and type%s", err)
	}
	if indexType != saelastic.CONNECTIVITYINDEXTYPE {
		t.Errorf("Excepected Index Type %s Got %s", saelastic.CONNECTIVITYINDEXTYPE, indexType)
	}
	if string(saelastic.CONNECTIVITYINDEX) != indexName {
		t.Errorf("Excepected Index %s Got %s", saelastic.CONNECTIVITYINDEX, indexName)
	}
}

func TestSanitize(t *testing.T) {
	result := saelastic.Sanitize(procEventDirty)
	log.Println(result)
}

func TestSanitizeOvs(t *testing.T) {
	result := saelastic.Sanitize(ovsEvent)
	log.Println(result)
}*/
