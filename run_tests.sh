#!/bin/bash
set -ex

# bootstrap
mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
yum install -y epel-release
yum install -y golang qpid-proton-c-devel iproute
go get -u golang.org/x/tools/cmd/cover
go get -u github.com/mattn/goveralls
go get -u golang.org/x/lint/golint
go get -u github.com/golang/dep/...
dep ensure -v --vendor-only

# run code validation tools
echo " *** Running pre-commit code validation"
./pre-commit

# run unit tests
echo " *** Running test suite"
# TODO: re-enable the test suite once supporting changes result in tests to pass
go test -v ./...

# check test coverage
echo " *** Running code coverage tooling"

# TODO: disable exiting on non-zero return because not all internal/pkg/*
#       contents have a corresponding Testing package
set +e

echo "mode: set" > coverage.out
for pkg in $(go list ./internal/pkg/...)
do
    go test -cover -coverpkg "$pkg" -coverprofile coverfragment.out ./tests/internal_pkg/* && grep -h -v "mode: set" coverfragment.out >> coverage.out
done

# upload coverage report
echo " *** Uploading coverage report to coveralls"
goveralls -coverprofile=coverage.out
