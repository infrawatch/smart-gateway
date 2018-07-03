package saelastic

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/redhat-nfvpe/telemetry-consumers/internal/pkg/elasticsearch/mapping"
)

//IndexName   ..
type IndexNameTest string

//COLLECTD
const (
	CONNECTIVITYINDEXTEST IndexNameTest = "collectd_connectivity_test"
	PROCEVENTINDEXTEST    IndexNameTest = "collectd_procevent_test"
	SYSEVENTINDEXTEST     IndexNameTest = "collectd_syslogs_test"
	GENERICINDEXTEST      IndexNameTest = "collectd_generic_test"
)

const connectivitydata = `[{"labels":{"alertname":"collectd_connectivity_gauge","instance":"d60b3c68f23e","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":2,"eventName":"interface eno2 up","lastEpochMicrosec":1518188764024922,"priority":"high","reportingEntityName":"collectd connectivity plugin","sequence":0,"sourceName":"eno2","startEpochMicrosec":1518188755700851,"version":1.0,"stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":1.0,"stateInterface":"eno2"}}},"startsAt":"2018-02-09T15:06:04.024859063Z"}]`

const connectivityDirty = "[{\"labels\":{\"alertname\":\"collectd_connectivity_gauge\",\"instance\":\"d60b3c68f23e\",\"connectivity\":\"eno2\",\"type\":\"interface_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"stateChange\\\",\\\"eventId\\\":11,\\\"eventName\\\":\\\"interface eno2 down\\\",\\\"lastEpochMicrosec\\\":1518790014024924,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd connectivity plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"eno2\\\",\\\"startEpochMicrosec\\\":1518790009881440,\\\"version\\\":1.0,\\\"stateChangeFields\\\":{\\\"newState\\\":\\\"outOfService\\\",\\\"oldState\\\":\\\"inService\\\",\\\"stateChangeFieldsVersion\\\":1.0,\\\"stateInterface\\\":\\\"eno2\\\"}}\"},\"startsAt\":\"2018-02-16T14:06:54.024856417Z\"}]"

const procEventData = `[{"labels":{"alertname":"collectd_procevent_gauge","instance":"d60b3c68f23e","procevent":"bla.py","type":"process_status","severity":"FAILURE","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"fault","eventId":3,"eventName":"process bla.py (8537) down","lastEpochMicrosec":1518791119579620,"priority":"high","reportingEntityName":"collectd procevent plugin","sequence":0,"sourceName":"bla.py","startEpochMicrosec":1518791111336973,"version":1.0,"faultFields":{"alarmCondition":"process bla.py (8537) state change","alarmInterfaceA":"bla.py","eventSeverity":"CRITICAL","eventSourceType":"process","faultFieldsVersion":1.0,"specificProblem":"process bla.py (8537) down","vfStatus":"Ready to terminate"}}},"startsAt":"2018-02-16T14:25:19.579573212Z"}]`

const procEventDirty = "[{\"labels\":{\"alertname\":\"collectd_procevent_gauge\",\"instance\":\"d60b3c68f23e\",\"procevent\":\"bla.py\",\"type\":\"process_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"fault\\\",\\\"eventId\\\":3,\\\"eventName\\\":\\\"process bla.py (8537) down\\\",\\\"lastEpochMicrosec\\\":1518791119579620,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd procevent plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"bla.py\\\",\\\"startEpochMicrosec\\\":1518791111336973,\\\"version\\\":1.0,\\\"faultFields\\\":{\\\"alarmCondition\\\":\\\"process bla.py (8537) state change\\\",\\\"alarmInterfaceA\\\":\\\"bla.py\\\",\\\"eventSeverity\\\":\\\"CRITICAL\\\",\\\"eventSourceType\\\":\\\"process\\\",\\\"faultFieldsVersion\\\":1.0,\\\"specificProblem\\\":\\\"process bla.py (8537) down\\\",\\\"vfStatus\\\":\\\"Ready to terminate\\\"}}\"},\"startsAt\":\"2018-02-16T14:25:19.579573212Z\"}]"

const ovsEvent = `[{"labels":{"alertname":"collectd_ovs_events_gauge","instance":"nfvha-comp-03","ovs_events":"br0","type":"link_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"link state of \"br0\" interface has been changed to \"UP\"","uuid":"c52f2aca-3cb1-48e3-bba3-100b54303d84"},"startsAt":"2018-02-22T20:12:19.547955618Z"}]`

const elastichost = "http://10.19.110.5:9200"

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	var client *ElasticClient
	client = CreateClient(elastichost, false, false)
	if client.err != nil {
		log.Fatalf("Failed to connect to elastic search%s", client.err)
	} else {
		client.TearDownTestIndex()
		os.Exit(m.Run())
	}
}

//InitAllMappings ....
func (ec *ElasticClient) TearDownTestIndex() {
	ec.DeleteIndex(string(CONNECTIVITYINDEXTEST))
	ec.DeleteIndex(string(PROCEVENTINDEXTEST))
	ec.DeleteIndex(string(SYSEVENTINDEXTEST))
	ec.DeleteIndex(string(GENERICINDEXTEST))
	//do not create now and leave it for defaults
	/*ec.CreateIndex(string(CONNECTIVITYINDEX), saelastic.ConnectivityMapping)
	ec.CreateIndex(string(PROCEVENTINDEX), saelastic.ConnectivityMapping)
	ec.CreateIndex(string(SYSEVENTINDEX), saelastic.ConnectivityMapping)
	*/
}

func TestIndexCheckConnectivity(t *testing.T) {
	indexName, indexType, err := GetIndexNameType(connectivitydata)
	if err != nil {
		t.Errorf("Failed to get indexname and type%s", err)
	}
	if indexType != CONNECTIVITYINDEXTYPE {
		t.Errorf("Excepected Index Type %s Got %s", CONNECTIVITYINDEXTYPE, indexType)
	}
	if string(CONNECTIVITYINDEX) != indexName {
		t.Errorf("Excepected Index %s Got %s", CONNECTIVITYINDEX, indexName)
	}

}
func TestSanitize(t *testing.T) {
	result := Sanitize(procEventDirty)
	log.Println(result)
}
func TestSanitizeOvs(t *testing.T) {
	result := Sanitize(ovsEvent)
	log.Println(result)
}

func TestClient(t *testing.T) {
	var client *ElasticClient
	client = CreateClient(elastichost, false, false)
	if client.err != nil {
		t.Errorf("Failed to connect to elastic search%s", client.err)
	}
}

func TestIndexCreateAndDelete(t *testing.T) {
	var client *ElasticClient
	client = CreateClient(elastichost, false, false)
	if client.err != nil {
		t.Errorf("Failed to connect to elastic search%s", client.err)
	} else {
		indexName, _, err := GetIndexNameType(connectivitydata)
		if err != nil {
			t.Errorf("Failed to get indexname and type%s", err)
			return
		}

		testIndexname := fmt.Sprintf("%s_%s", indexName, "test")
		client.DeleteIndex(testIndexname)

		client.CreateIndex(testIndexname, saelastic.ConnectivityMapping)
		exists, err := client.client.IndexExists(string(testIndexname)).Do(client.ctx)
		if exists == false || err != nil {
			t.Errorf("Failed to create index %s", err)
		}
		err = client.DeleteIndex(testIndexname)
		if err != nil {
			t.Errorf("Failed to Delete index %s", err)
		}
	}

}

func TestConnectivityDataCreate(t *testing.T) {
	var client *ElasticClient
	client = CreateClient(elastichost, false, false)
	if client.err != nil {
		t.Errorf("Failed to connect to elastic search%s", client.err)
	} else {
		indexName, IndexType, err := GetIndexNameType(connectivitydata)
		if err != nil {
			t.Errorf("Failed to get indexname and type%s", err)
			return
		}
		testIndexname := fmt.Sprintf("%s_%s", indexName, "test")
		err = client.DeleteIndex(testIndexname)

		client.CreateIndex(testIndexname, saelastic.ConnectivityMapping)
		exists, err := client.client.IndexExists(string(testIndexname)).Do(client.ctx)
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
	}

}
