#!/bin/bash
set -ex


# bootstrap
mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
sed -i '/^tsflags=.*/a ip_resolve=4' /etc/yum.conf
yum install -y epel-release
yum install -y git golang qpid-proton-c-devel iproute
go get -u golang.org/x/tools/cmd/cover
GO111MODULE=off go get -u github.com/mattn/goveralls
go get -u golang.org/x/lint/golint
go get -u honnef.co/go/tools/cmd/staticcheck

# get vendor code
go mod vendor

# run code validation tools
echo " *** Running pre-commit code validation"
./build/test-framework/pre-commit

# run unit tests
echo " *** Running test suite"
# TODO: re-enable the test suite once supporting changes result in tests to pass
go test -v ./...

# check test coverage
echo " *** Running code coverage tooling"
go test ./... -race -covermode=atomic -coverprofile=coverage.txt

# upload coverage report
echo " *** Uploading coverage report to coveralls"
goveralls -coverprofile=coverage.txt
