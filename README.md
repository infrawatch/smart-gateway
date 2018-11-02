# smart-gateway ![build status](https://travis-ci.org/redhat-nfvpe/smart-gateway.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/redhat-nfvpe/smart-gateway)](https://goreportcard.com/report/github.com/redhat-nfvpe/smart-gateway)

Smart gateway for service assurance. Includes applications for both metrics and events gathering.

Provides middleware that connects to an AMQP 1.0 message bus, pulling data off the bus and exposing it as a 
scrape target for Prometheus. Metrics are provided via the OPNFV Barometer project (collectd). Events are
provided by the various event plugins for collectd, including connectivity, procevent and sysevent.

# Dependencies

Dependencies are managed using [`dep`](https://github.com/golang/dep). Clone this project, then obtain the
dependencies with the following commands.

```
go get -u github.com/redhat-nfvpe/smart-gateway
cd $GOPATH/src/github.com/redhat-nfvpe/smart-gateway
dep ensure -v -vendor-only
```

# Building

Build the `smart_gateway` with Golang using the following command.

```
cd $GOPATH/src/github.com/redhat-nfvpe/smart-gateway
go build -o smart_gateway cmd/main.go 
```
