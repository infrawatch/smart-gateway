#!/bin/bash

set -x

mkdir -p /go/bin /go/src /go/pkg
export GOPATH=/go
export PATH=$PATH:$GOPATH/bin

# get dependencies
yum install -y epel-release
yum install -y golang qpid-proton-c-devel iproute
go get -u github.com/golang/dep
sh $GOPATH/src/github.com/golang/dep/install.sh

ss -lnp
ss -ntp
# run test suite
dep ensure -v --vendor-only
go test -v -timeout=10s ./tests/*
