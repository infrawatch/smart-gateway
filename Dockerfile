FROM fedora:27 AS builder
ENV GOPATH=/go
ENV D=/go/src/github.com/redhat-nfvpe/smart-gateway

RUN dnf install qpid-proton-c-devel git golang -y && \
        dnf clean all && \
        go get -u github.com/golang/dep/...

WORKDIR $D
COPY . $D/
RUN /go/bin/dep ensure -v -vendor-only && \
        go build -o smart_gateway cmd/main.go && \
        cp smart_gateway /tmp/

FROM fedora:27
LABEL maintainer="admin@nfvpe.site"
RUN dnf install qpid-proton-c -y && dnf clean all
COPY --from=builder /tmp/smart_gateway /
EXPOSE 8081
EXPOSE 5672
ENTRYPOINT ["/smart_gateway"]
