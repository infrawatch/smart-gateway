#!/bin/bash
set -ex


# bootstrap
mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
sed -i '/^tsflags=.*/a ip_resolve=4' /etc/yum.conf
yum install -y epel-release
# below is not available currently
#yum install -y https://centos7.iuscommunity.org/ius-release.rpm
#yum remove -y git*
#yum install -y git216-all
yum install -y golang qpid-proton-c-devel iproute
go get -u golang.org/x/tools/cmd/cover
go get -u github.com/mattn/goveralls
go get -u golang.org/x/lint/golint
go get -u github.com/golang/dep/...
go get -u honnef.co/go/tools/cmd/staticcheck
dep ensure -v --vendor-only

# run code validation tools
echo " *** Running pre-commit code validation"
./build/test-framework/pre-commit

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
goveralls -service=travis-ci -coverprofile=coverage.out
