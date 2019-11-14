# smart-gateway ![build status](https://travis-ci.org/redhat-service-assurance/smart-gateway.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/redhat-service-assurance/smart-gateway)](https://goreportcard.com/report/github.com/redhat-service-assurance/smart-gateway) [![Coverage Status](https://coveralls.io/repos/github/redhat-service-assurance/smart-gateway/badge.svg)](https://coveralls.io/github/redhat-service-assurance/smart-gateway) [![Docker Repository on Quay](https://quay.io/repository/redhat-service-assurance/smart-gateway/status "Docker Repository on Quay")](https://quay.io/repository/redhat-service-assurance/smart-gateway)

Smart gateway for service assurance. Includes applications for both metrics and
events gathering.

Provides middleware that connects to an AMQP 1.0 message bus, pulling data off
the bus and exposing it as a scrape target for Prometheus. Metrics are provided
via the OPNFV Barometer project (collectd). Events are provided by the various
event plugins for collectd, including connectivity, procevent and sysevent.

# Dependencies

Dependencies are managed using [`dep`](https://github.com/golang/dep). Clone
this project, then obtain the dependencies with the following commands. Example
below is built on CentOS 7.

```
go get -u github.com/redhat-service-assurance/smart-gateway
go get -u github.com/golang/dep/...
yum install -y epel-release
yum install -y golang qpid-proton-c-devel iproute
cd $GOPATH/src/github.com/redhat-service-assurance/smart-gateway
dep ensure -v -vendor-only
```

# Building Smart Gateway

## Building with Golang

Build the `smart_gateway` with Golang using the following command.

```
cd $GOPATH/src/github.com/redhat-service-assurance/smart-gateway
go build -o smart_gateway cmd/main.go
```

# Building with Docker

Building the `smart-gateway` with docker using the following commands.

```
git clone --depth=1 --branch=master https://github.com/redhat-service-assurance/smart-gateway.git smart-gateway; rm -rf ./smart-gateway/.git
cd smart-gateway
sudo docker build -t smart-gateway --file=build/Dockerfile .
```

# Testing

The Smart Gateway ships with various unit tests located in the `tests/`
directory. To execute these unit tests, run the following command from the
top-level directory.

```
go test -v ./...
```

> **A note about test layout in Smart Gateway**
> 
> Generally tests are shipped in Golang directly within the packages as
> `<implementation>_test.go`, for example, `alerts.go` will have a
> corresponding `alerts_test.go` within the `alerts` package.
> 
> In the Smart Gateway, we've purposely taken a separate approach by moving the
> tests into their own package, located within the `tests/` subdirectory. The
> reason for this is two-fold:
> 
> 1. It is the recommended workaround for avoiding falling into cyclic
>    dependencies / ciclical import problems.
> 1. Test implementer loses access to private logic, so has to think about the
>    tested API more in depth. Also when the test of package A breaks, we can
>    be sure that other packages are broken too. On the other side, when you
>    have access to private logic, you can unintentionally workaround issues
>    and hit the bugs in production deployments.
