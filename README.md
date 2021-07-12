# smart-gateway ![build status](https://travis-ci.org/infrawatch/smart-gateway.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/infrawatch/smart-gateway)](https://goreportcard.com/report/github.com/infrawatch/smart-gateway) [![Coverage Status](https://coveralls.io/repos/github/infrawatch/smart-gateway/badge.svg)](https://coveralls.io/github/infrawatch/smart-gateway) [![Docker Repository on Quay](https://quay.io/repository/infrawatch/smart-gateway/status "Docker Repository on Quay")](https://quay.io/repository/infrawatch/smart-gateway)

Smart Gateway for Service Telemetry Framework. Includes applications for both
metrics and events gathering.

> **WARNING** 
> 
> This repository contains the legacy Smart Gateway, no longer in use as of STF 1.3 and will no longer be released once STF 1.4 is available and STF 1.2 moves to EOL. If this code base looks interesting or useful for your project, you are strongly encouraged to use sg-bridge and sg-core which have significantly better performance and integration capabilities. See https://github.com/infrawatch/sg-bridge and https://github.com/infrawatch/sg-core

Provides middleware that connects to an AMQP 1.0 message bus, pulling data off
the bus and exposing it as a scrape target for Prometheus. Metrics are provided
via the OPNFV Barometer project (collectd) and Ceilometer (OpenStack). Events
are provided by the various event plugins for collectd, including connectivity,
procevent and sysevent, and Ceilometer.

# Dependencies

Dependencies are managed using golang modules. Clone this project, then obtain
the dependencies with the following commands. Example below is built on CentOS
8.

```
go get -u github.com/infrawatch/smart-gateway
sudo curl -L https://trunk.rdoproject.org/centos8-master/delorean-deps.repo -o /etc/yum.repos.d/delorean-deps.repo
dnf install -y golang qpid-proton-c-devel iproute
cd $GOPATH/src/github.com/infrawatch/smart-gateway
```

# Building Smart Gateway

## Building with Golang

Build the `smart_gateway` with Golang using the following command.

```
cd $GOPATH/src/github.com/infrawatch/smart-gateway
go build -o smart_gateway cmd/main.go
```

# Building with Docker

Building the `smart-gateway` with docker using the following commands.

```
git clone --depth=1 --branch=master https://github.com/infrawatch/smart-gateway.git smart-gateway; rm -rf ./smart-gateway/.git
cd smart-gateway
docker build -t smart-gateway .
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
