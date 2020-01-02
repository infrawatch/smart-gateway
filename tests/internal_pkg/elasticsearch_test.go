package tests

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saconfig"
	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic"
)

//COLLECTD
const (
	elastichost = "http://127.0.0.1:9200"
	testCACert  = `-----BEGIN CERTIFICATE-----
MIIDSTCCAjGgAwIBAgIUVLbF9klC/t0fQoG35GAVTjU6tYEwDQYJKoZIhvcNAQEL
BQAwNDEyMDAGA1UEAxMpRWxhc3RpYyBDZXJ0aWZpY2F0ZSBUb29sIEF1dG9nZW5l
cmF0ZWQgQ0EwHhcNMTkwOTAzMDkwNTUxWhcNMjIwOTAyMDkwNTUxWjA0MTIwMAYD
VQQDEylFbGFzdGljIENlcnRpZmljYXRlIFRvb2wgQXV0b2dlbmVyYXRlZCBDQTCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKKOQqQFVvlBqFc9K9ESM49+
RFqNXdeStK+sVkZ9WkvmfSfj5h91O9BXev88n9dqcifmbS99KiT6ExzX3RO1NDxq
mIHGiscaalYA7gJlbF90cqvuy4ejNs50DDgSAeDLTHEn+q5PJeY7uQweQQ1usnFR
DbevOH/ubjdNRlTlockl1iYd8voQoRNxCgeN8JKd1XDyXXQm+sdZP87hnMgfDj4A
r88TkhbXTFhtWcU7aLi/uNq0u/3CfJwkwvH7SFuqv/qnqXXu+7vaA+zifGSHmIMS
GX47Ki4ordGv75hFs70gI3qtgq5Ce1+4sGl05Ime/4+iRoj2S/EKrbSejnOklgMC
AwEAAaNTMFEwHQYDVR0OBBYEFCvqtlWPfEyQCOus3n+NjVJrmsYdMB8GA1UdIwQY
MBaAFCvqtlWPfEyQCOus3n+NjVJrmsYdMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggEBABPy/tMJypO4TIEakRfUAjo23za3DSH4aN9FjuF5dOnBAKU3
6Wxf2abDwaTUTh/wnuBrh8ubuFWQyqCEL8+ncxjgeEpOHpvbxrnVfFQxDt7rdAqK
VRGddwUCaHgJ1ZBdhrLuSmWwaXsQL4q2F4dLifq/BIdOPvT3lHzPh/D5sdCcPVrX
V2j6pIReP/TfM+7NIlLSL+xPTjMV1lTFMupYrZDUouB5lkqyNgO0/eXcBPjFjdVz
5Kx1xUfPcx8oSotFlrqA4eXfeQBFr9dJDsTeEZNSUM41TQKRoPn4qdPNQ/QPoJgR
Mig5sWoQl+8PDYeSCcgmmWF/uPpAt9bORvtmj8U=
-----END CERTIFICATE-----`
	testClientCert = `-----BEGIN CERTIFICATE-----
MIIDTjCCAjagAwIBAgIUC1CKg5RQAEHSl672tLWVHwQ6UCswDQYJKoZIhvcNAQEL
BQAwNDEyMDAGA1UEAxMpRWxhc3RpYyBDZXJ0aWZpY2F0ZSBUb29sIEF1dG9nZW5l
cmF0ZWQgQ0EwHhcNMTkwOTAzMDkwNTUyWhcNMjIwOTAyMDkwNTUyWjARMQ8wDQYD
VQQDEwZub2RlLTEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCMBgck
nkC4ZcRSsrvw/TSLUJ+Fnzox9mHmvawItfhIjrhPpcz8kgEQ2NvTSFQ5i6mt3wca
bUqCRrqJ7HZ4Lk4epPbL50GYFn/I98oBqI6SH3I/At6ZnTUcUgwZCaWZ8iHrQ3Bd
EP+LWuAIRs1IH9Kg/+uP7q1zhb3yEUx84PBNNVNC10i5w1Gtd2LsgQis8mA2zLG5
IjVeAyLe1zyc4oM74TxULr+vRv5gZFGJMbO9FXq/ztNOwv1YQ2RatNY3aEk/NMBj
pUuGuCxMdTcU5/sOtaLIroaCR6BNNe1B3RxnBuqyxvmwwk+RlXchqPtMWEW7XDBI
tO/jLSC/zkbD4yEnAgMBAAGjezB5MB0GA1UdDgQWBBS1aF7Zhl3xRhkkWsimErYf
9gaH+DAfBgNVHSMEGDAWgBQr6rZVj3xMkAjrrN5/jY1Sa5rGHTAsBgNVHREEJTAj
gglsb2NhbGhvc3SHBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwCQYDVR0TBAIwADAN
BgkqhkiG9w0BAQsFAAOCAQEAEqprw19/7A22xSxlwxDgpB7aE7Cxyn3GfMxsb3vE
6h+oIoCwEHTKPqeJLMF/SnLLjdqRTQ43nU2bKpjfTJ8lSzX6ccNWVoMKMUNkSBkU
FMmR8e/gaqWTPiRqcSJfuVwG4L06F7wcyHSqBgkJBErdttWHbFXmYdhleui7xDg5
whi8l6c7TS2qMuLo1JnvvyfoEvxuo8RKvji11t+ZuSrXp0fq9dFQEgnzAoekLutO
ygoZsqvrMRK2F0U4XS9e2JGyLMOz0oxvUtZMRVFVtR5AUpmzdz42LGWnT4xxL6jO
vB6iVwxM7ZjAGrAJg8hOTvSTn/0X5HqCCNwrQ2tQyfuZ7g==
-----END CERTIFICATE-----`
	testClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAjAYHJJ5AuGXEUrK78P00i1CfhZ86MfZh5r2sCLX4SI64T6XM
/JIBENjb00hUOYuprd8HGm1Kgka6iex2eC5OHqT2y+dBmBZ/yPfKAaiOkh9yPwLe
mZ01HFIMGQmlmfIh60NwXRD/i1rgCEbNSB/SoP/rj+6tc4W98hFMfODwTTVTQtdI
ucNRrXdi7IEIrPJgNsyxuSI1XgMi3tc8nOKDO+E8VC6/r0b+YGRRiTGzvRV6v87T
TsL9WENkWrTWN2hJPzTAY6VLhrgsTHU3FOf7DrWiyK6GgkegTTXtQd0cZwbqssb5
sMJPkZV3Iaj7TFhFu1wwSLTv4y0gv85Gw+MhJwIDAQABAoIBABGWBDmmIozGQ0T7
q70VoA7LPm3C1MVHo34eXkftytQaEK34Leme0MFz6w/7KpDbqKDsvPClv1DjXzRJ
XYu0jR0uLMzpK4TVdpEgBd/1coqJpoihbKGwa+Y1q81NN95A2d+5ZZhatS2kaTTA
57FiRcrwuX4nROOYbYXEhG2+to+LrkpEGqG+wgroWMuClfFcPxfnp5thiX1lcP5X
t1L9IsbHrpxGA5HO+4gtLAmgtA2OVXZ6R2eJVlrFbhrfqoN1JiixrWLoCrbS1vpn
dPLG0bOpm6RqiH627+hUVLSYtKNtj9T8FbWlHXCMH19MXG1DwTPaI+RJCb3AasFl
saIKyzkCgYEAzEwy/odj5mrWIWCBnUSkwBUHCxSFtvrGdzwcVJA/2CGyaF49Hkkc
HM2c4v1QjVlSn5niqr8EyKcU/pbq83fUiO7AT8qH9JDlyWCkuslJ/C22/KhGnzFl
g5f4UXr8/DoLP05a+cP5AorhhuxRTjxTUD5lMk81ipXgnGp4jytHOSsCgYEAr3W2
RRnKQr6u8zyVKUlR31OLAgT16nAVkC3rl2FPyf02d4bz5+Gv7LPxYbUg8zAWvMR1
ArdYAh3zjnfAYzoBPfCakXxc9Hbwl7GMBOs9UOEz6sSOcQ8G6t/Xx5oLeXcXvxgJ
whpLPfu8zgucqy1PzoeXTKY1dzEthgy6nGPgwfUCgYEAmS75fYgfDAJHlLc7+KQj
tDMQGOrGaDEY5waXZ4DRnkmF8GPZCABhp+c0H6842wOCxFEqeETKXXmKcGrQuMW9
Av+iCzIdRu/unFRur++GHiRY9JFogq0TJNyqQM4rKySKkmk6JdUfvRxNhlFjlXn+
LkjasCJcTxGaXS4oP5F/0gkCgYAnkpfqW9e3WARjTa2ioyu4/8GhUfcYyfDDFOhG
uybguqBXMvO9v7QK4ca2L8DfuF/YcUKmuy05RQIShsW4W3O+QY7K806PwGeg/uVC
kr/AhxpLf8tUinwX6yZimUavPYH4knZY9c80iptZqVrLbKvMO96O5gm2+Tt4OVS5
QvmFJQKBgQCKmrVQ9at1oNwzPiEIlQszyZ9n5vrxi1EbpnRZjAv9KBErjLVJOEA7
u+Jpmr1o0z9CPvXFmdWGdF2dJrgBImgQnlsNVctK8x1m0azfduPcgPgKSlnMdnRS
wg4Luw64Vn3osASCHv5gwoIgBepLpOby7KrCEvOwuFyB9QGZXXIxBQ==
-----END RSA PRIVATE KEY-----`
)

/*
func TestMain(t *testing.T) {
	config := saconfig.EventConfiguration{
		Debug:          false,
		ElasticHostURL: elastichost,
		UseTLS:         false,
		TLSClientCert:  "",
		TLSClientKey:   "",
		TLSCaCert:      "",
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
*/

func TestTls(t *testing.T) {
	dir, err := ioutil.TempDir("", "sg-test-tls")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(dir)

	verifyConnection := true
	clientCert := os.Getenv("SA_TESTS_ES_CLIENT_CERT")
	if len(clientCert) == 0 {
		verifyConnection = false
		clientCert = path.Join(dir, "client.cert")
		err = ioutil.WriteFile(clientCert, []byte(testClientCert), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary client cert: %s", err)
		}
	}
	clientKey := os.Getenv("SA_TESTS_ES_CLIENT_KEY")
	if len(clientKey) == 0 {
		verifyConnection = false
		clientKey = path.Join(dir, "client.key")
		err = ioutil.WriteFile(clientKey, []byte(testClientKey), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary client key: %s", err)
		}
	}
	caCert := os.Getenv("SA_TESTS_ES_CA_CERT")
	if len(caCert) == 0 {
		verifyConnection = false
		caCert = path.Join(dir, "ca.cert")
		err = ioutil.WriteFile(caCert, []byte(testCACert), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary ca cert: %s", err)
		}
	}

	t.Run("Test insecure connection", func(t *testing.T) {
		config := saconfig.EventConfiguration{
			Debug:          false,
			ElasticHostURL: elastichost,
			UseTLS:         true,
			TLSClientCert:  clientCert,
			TLSClientKey:   clientKey,
			TLSCaCert:      caCert,
			TLSServerName:  "",
		}

		_, err = saelastic.CreateClient(config)
		if err != nil && verifyConnection {
			t.Fatalf("Failed to connect to elastic search using TLS: %s", err)
		}
	})

	t.Run("Test unset ServerName", func(t *testing.T) {
		config := saconfig.EventConfiguration{
			Debug:          false,
			ElasticHostURL: elastichost,
			UseTLS:         true,
			TLSClientCert:  clientCert,
			TLSClientKey:   clientKey,
			TLSCaCert:      caCert,
		}

		_, err = saelastic.CreateClient(config)
		if err != nil && verifyConnection {
			t.Fatalf("Failed to connect to elastic search using TLS: %s", err)
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
}*/
