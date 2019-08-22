package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/olivere/elastic"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic"
)

//ElasticClient  ....
type ElasticClient struct {
	client *elastic.Client
	ctx    context.Context
	err    error
}

// createTlsClient creates http.Client for elastic.Client with enabled
// cert-based authentication
func createTlsClient(serverName string, certFile string, keyFile string, caFile string) (*http.Client, error) {
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
		ServerName:   serverName,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
	tlsConfig.BuildNameToCertificate()
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}, nil
}

//CreateClient   ....
func CreateClient(config saconfig.EventConfiguration) (*ElasticClient, error) {
	var elasticClient *ElasticClient
	elasticOpts := []elastic.ClientOptionFunc{elastic.SetHealthcheckInterval(5 * time.Second), elastic.SetURL(config.ElasticHostURL)}
	// add transport with TLS enabled in case it is required
	if config.UseTls {
		tlsClient, err := createTlsClient(config.TlsServerName, config.TlsClientCert, config.TlsClientKey, config.TlsCaCert)
		if err != nil {
			return elasticClient, nil
		}
		elasticOpts = append(elasticOpts, elastic.SetHttpClient(tlsClient))
	}

	eclient, err := elastic.NewClient(elasticOpts...)
	if err != nil {
		log.Fatal(err)
		return elasticClient, err
	}
	elasticClient = &ElasticClient{client: eclient, ctx: context.Background()}
	return elasticClient, nil
}

func (ec *ElasticClient) IndexExists(index string) *elastic.IndicesExistsService {
	return ec.client.IndexExists(index)
}

func main() {
	fConfigLocation := flag.String("config", "", "Path to configuration file.")
	flag.Parse()

	serverConfig := saconfig.LoadEventConfig(*fConfigLocation)

	if len(serverConfig.ElasticHostURL) == 0 {
		log.Println("Elastic Host URL is required")
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("caught sig: %+v", sig)
			log.Println("Wait for 2 second to finish processing")
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
	}()

	log.Printf("Connecting to ElasticSearch : %s\n", serverConfig.ElasticHostURL)
	elasticClient, _ := saelastic.CreateClient(serverConfig)

	log.Printf("Existing index search: %v\n", elasticClient.IndexExists("test"))
	log.Printf("Non-existing index search: %v\n", elasticClient.IndexExists("foobar"))
}
