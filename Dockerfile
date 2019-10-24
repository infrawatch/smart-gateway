# --- build smart gateway ---
FROM centos:7 AS builder
ENV GOPATH=/go
ENV D=/go/src/github.com/redhat-service-assurance/smart-gateway

WORKDIR $D
COPY . $D/

RUN yum install epel-release -y && \
        yum update -y --setopt=tsflags=nodocs && \
        yum install qpid-proton-c-devel git golang --setopt=tsflags=nodocs -y && \
        yum clean all && \
        go get -u github.com/golang/dep/... && \
        /go/bin/dep ensure -v -vendor-only && \
        go build -o smart_gateway cmd/main.go && \
        mv smart_gateway /tmp/

# --- end build, create smart gateway layer ---
FROM centos:7

LABEL io.k8s.display-name="Service Assurance Smart Gateway" \
      io.k8s.description="A component of the Service Assurance Framework on the server side that ingests data from AMQP 1.x and provides a metrics scrape endpoint for Prometheus, and forwards events to ElasticSearch" \
      maintainer="Leif Madsen <leif@redhat.com>"

RUN yum install epel-release -y && \
        yum update -y --setopt=tsflags=nodocs && \
        yum install qpid-proton-c --setopt=tsflags=nodocs -y && \
        yum clean all && \
        rm -rf /var/cache/yum

COPY --from=builder /tmp/smart_gateway /

ENTRYPOINT ["/smart_gateway"]
