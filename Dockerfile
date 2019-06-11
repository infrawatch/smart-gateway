# --- build smart gateway ---
FROM registry.access.redhat.com/ubi7/ubi AS builder
USER root
ENV GOPATH=/go
ENV D=/go/src/github.com/redhat-service-assurance/smart-gateway

WORKDIR $D
COPY . $D/

RUN yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y && \
yum install https://www.mercurial-scm.org/release/centos7/RPMS/x86_64/mercurial-3.4-0.x86_64.rpm -y && \
yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/libserf-1.3.9-1.el7.x86_64.rpm -y && \
yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/subversion-1.11.0-1.x86_64.rpm -y && \
        yum update -y --setopt=tsflags=nodocs && \
        yum install qpid-proton-c-devel git golang --setopt=tsflags=nodocs -y && \
        yum clean all && \
        go get -u github.com/golang/dep/... && \
        /go/bin/dep ensure -v -vendor-only && \
        go build -o smart_gateway cmd/main.go && \
        mv smart_gateway /tmp/

# --- end build, create smart gateway layer ---
FROM registry.access.redhat.com/ubi7/ubi

LABEL io.k8s.display-name="Service Assurance Smart Gateway" \
      io.k8s.description="A component of the Service Assurance Framework on the server side that ingests data from AMQP 1.x and provides a metrics scrape endpoint for Prometheus, and forwards events to ElasticSearch" \
      maintainer="Leif Madsen <leif@redhat.com>"

RUN yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm -y && \
yum install https://www.mercurial-scm.org/release/centos7/RPMS/x86_64/mercurial-3.4-0.x86_64.rpm -y && \
yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/libserf-1.3.9-1.el7.x86_64.rpm -y && \
yum install http://opensource.wandisco.com/centos/7/svn-1.11/RPMS/x86_64/subversion-1.11.0-1.x86_64.rpm -y && \
        yum update -y --setopt=tsflags=nodocs && \
        yum install qpid-proton-c --setopt=tsflags=nodocs -y && \
        yum clean all && \
        rm -rf /var/cache/yum

COPY --from=builder /tmp/smart_gateway /

ENTRYPOINT ["/smart_gateway"]
