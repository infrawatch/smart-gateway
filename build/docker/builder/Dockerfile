FROM fedora:27
ENV GOPATH=/go
ENV D=/go/src/github.com/redhat-nfvpe/telemetry-consumers

RUN dnf install qpid-proton-c-devel git golang -y && \
        dnf clean all && \
        go get -u github.com/golang/dep/...

WORKDIR $D
ONBUILD COPY . $D/
ONBUILD RUN /go/bin/dep ensure -v -vendor-only && \
        go build -o telemetry_consumer cmd/main.go && \
        cp telemetry_consumer /tmp/ 