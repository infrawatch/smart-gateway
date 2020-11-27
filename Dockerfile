# --- build smart gateway ---
FROM registry.access.redhat.com/ubi8 AS builder
ENV GOPATH=/go
ENV D=/go/src/github.com/infrawatch/smart-gateway

WORKDIR $D
COPY . $D/
COPY build/repos/opstools.repo /etc/yum.repos.d/opstools.repo

RUN dnf install qpid-proton-c-devel git golang --setopt=tsflags=nodocs -y && \
    go build -o smart_gateway cmd/main.go && \
    mv smart_gateway /tmp/

# --- end build, create smart gateway layer ---
FROM registry.access.redhat.com/ubi8

COPY build/repos/opstools.repo /etc/yum.repos.d/opstools.repo

RUN dnf install qpid-proton-c --setopt=tsflags=nodocs -y && \
    dnf clean all && \
    rm -rf /var/cache/yum

COPY --from=builder /tmp/smart_gateway /

ENTRYPOINT ["/smart_gateway"]

LABEL io.k8s.display-name="Smart Gateway" \
      io.k8s.description="A component of the Service Telemetry Framework on the server side that ingests data from AMQP 1.x and provides a metrics scrape endpoint for Prometheus, and forwards events to ElasticSearch" \
      maintainer="Leif Madsen <leif+smartgateway@redhat.com>"
