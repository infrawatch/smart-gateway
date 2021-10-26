package saelastic

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/olivere/elastic"
)

var debuges = func(format string, data ...interface{}) {} // Default no debugging output

//ElasticClient  ....
type ElasticClient struct {
	client *elastic.Client
	ctx    context.Context
}

//InitAllMappings removes all indices with prefixes used by smart-gateway
func (ec *ElasticClient) InitAllMappings() {
	ec.DeleteIndex(string("collectd_*"))
	ec.DeleteIndex(string("ceilometer_*"))
	ec.DeleteIndex(string("generic_*"))
}

//createTLSClient creates http.Client for elastic.Client with enabled
//cert-based authentication
func createTLSClient(serverName string, certFile string, keyFile string, caFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
		return &http.Client{}, err
	}

	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal(err)
		return &http.Client{}, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(ca)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
	if len(serverName) == 0 {
		tlsConfig.InsecureSkipVerify = true
	} else {
		tlsConfig.ServerName = serverName
	}
	debuges("InsecureSkipVerify is set to %t", tlsConfig.InsecureSkipVerify)

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}, nil
}

//CreateClient   ....
func CreateClient(config saconfig.EventConfiguration) (*ElasticClient, error) {
	if config.Debug {
		debuges = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	var elasticClient *ElasticClient
	elasticOpts := []elastic.ClientOptionFunc{elastic.SetHealthcheckInterval(5 * time.Second), elastic.SetURL(config.ElasticHostURL)}
	// add transport with TLS enabled in case it is required
	if config.UseTLS {
		tlsClient, err := createTLSClient(config.TLSServerName, config.TLSClientCert, config.TLSClientKey, config.TLSCaCert)
		if err != nil {
			return elasticClient, nil
		}
		elasticOpts = append(elasticOpts, elastic.SetHttpClient(tlsClient), elastic.SetScheme("https"))
	}

	if config.UseBasicAuth {
		elasticOpts = append(elasticOpts, elastic.SetBasicAuth(config.ElasticUser, config.ElasticPass))
	}

	eclient, err := elastic.NewClient(elasticOpts...)
	if err != nil {
		return elasticClient, err
	}
	elasticClient = &ElasticClient{client: eclient, ctx: context.Background()}
	if config.ResetIndex {
		elasticClient.InitAllMappings()
	}
	debuges("Debug:ElasticSearch client created.")
	return elasticClient, nil
}

//IndexExists ...
func (ec *ElasticClient) IndexExists(index string) *elastic.IndicesExistsService {
	return ec.client.IndexExists(index)
}

//GetContext ...
func (ec *ElasticClient) GetContext() context.Context {
	return ec.ctx
}

//CreateIndex  ...
func (ec *ElasticClient) CreateIndex(index string, mapping string) {

	exists, err := ec.client.IndexExists(string(index)).Do(ec.ctx)
	if err != nil {
		// Handle error nothing to do index exists
		debuges("Debug:ElasticSearch indexExists returned an error: %s", err)
	}
	if !exists {
		// Index does not exist yet.
		// Create a new index.
		createIndex, err := ec.client.CreateIndex(string(index)).BodyString(mapping).Do(ec.ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			log.Println("Index Not acknowledged")
		}
	}

}

//genUUIDv4  ...
func genUUIDv4() string {
	id, _ := uuid.NewV4()
	debuges("Debug:github.com/satori/go.uuid:   %s\n", id)
	return id.String()
}

// Generate an id based on the data itself to prevent duplicate events from multiple (HA) instances of the SG
func genHashedID(jsondata interface{}) string {
	dataBytes, err := json.Marshal(jsondata)
	if err != nil {
		log.Printf("Error during event encoding: %v\n", err)
		uuid := genUUIDv4()
		log.Printf("Will return a UUID string instead (duplicate events may occur): %v\n", uuid)
		return uuid
	}
	eventHashBytes := sha1.Sum(dataBytes)
	eventHashHex := hex.EncodeToString(eventHashBytes[:])
	return eventHashHex
}

//Create ...  it can be BodyJson or BodyString.. BodyJson needs struct defined
func (ec *ElasticClient) Create(indexname string, indextype string, jsondata interface{}) (string, error) {
	ctx := ec.ctx
	id := genHashedID(jsondata)

	debuges("Debug:Printing body %s\n", jsondata)
	result, err := ec.client.Index().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		BodyJson(jsondata).
		Do(ctx)
	if err != nil {
		// Handle error
		debuges("Create document Error %#v", err)
		return id, err
	}
	debuges("Debug:Indexed  %s to index %s, type %s\n", result.Id, result.Index, result.Type)
	// Flush to make sure the documents got written.
	// Flush asks Elasticsearch to free memory from the index and
	// flush data to disk.
	_, err = ec.client.Flush().Index(string(indexname)).Do(ctx)
	return id, err

}

//DeleteIndex ...
func (ec *ElasticClient) DeleteIndex(index string) error {
	// Delete an index.
	deleteIndex, err := ec.client.DeleteIndex(string(index)).Do(ec.ctx)
	if err != nil {
		// Handle error
		//panic(err)
		return err
	}
	if !deleteIndex.Acknowledged {
		debuges("Debug:ElasticSearch DeleteIndex not acknowledged")
	}
	return nil
}

//Delete  ....
func (ec *ElasticClient) Delete(indexname string, indextype string, id string) error {
	_, err := ec.client.Delete().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
	return err
}

//Get  ....
func (ec *ElasticClient) Get(indexname string, indextype string, id string) (*elastic.GetResult, error) {
	result, err := ec.client.Get().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
	if err != nil {
		// Handle error
		return nil, err
	}
	if result.Found {
		debuges("Debug:Got document %s in version %d from index %s, type %s\n", result.Id, result.Version, result.Index, result.Type)
	}
	return result, nil
}

//Search  ..
func (ec *ElasticClient) Search(indexname string) *elastic.SearchResult {
	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := ec.client.Search().
		Index(indexname).   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ec.ctx)          // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	debuges("Debug:Query took %d milliseconds\n", searchResult.TookInMillis)
	return searchResult
}
