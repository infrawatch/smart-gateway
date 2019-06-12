#!/bin/bash

set -x

mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
#yum install -y epel-release

#yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y
#yum install https://www.mercurial-scm.org/release/centos7/RPMS/x86_64/mercurial-3.4-0.x86_64.rpm -y 
#yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/libserf-1.3.9-1.el7.x86_64.rpm
#yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/subversion-1.11.0-1.x86_64.rpm -y

yum --nogpgcheck localinstall ./repo/epel-release-latest-7.noarch.rpm -y
yum --nogpgcheck localinstall ./repo/libserf-1.3.9-1.el7.x86_64.rpm -y
yum --nogpgcheck localinstall ./repo/mercurial-3.4-0.x86_64.rpm -y
yum --nogpgcheck localinstall ./repo/subversion-1.11.0-1.x86_64.rpm -y

yum install -y golang qpid-proton-c-devel iproute
go get -u golang.org/x/tools/cmd/cover
go get -u github.com/mattn/goveralls
go get -u golang.org/x/lint/golint
go get -u github.com/golang/dep
sh $GOPATH/src/github.com/golang/dep/install.sh

# run test suite
dep ensure -v --vendor-only
go test -v -timeout=10s ./tests/*

# run lints
go vet ./cmd/main.go
golint . | xargs -r false

# analyze test coverage
# works correctly only with Go-1.11+ due to: https://github.com/golang/go/issues/25093
echo "mode: set" > coverage.out
for pkg in $(go list ./internal/pkg/...); do go test -cover -coverpkg $pkg -coverprofile coverfragment.out ./tests/internal_pkg/* && grep -h -v "mode: set" coverfragment.out >> coverage.out; done
$GOPATH/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $COVERALLS_TOKEN || exit 0
