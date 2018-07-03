FROM nfvpe/smart_agent_builder:latest AS builder

FROM fedora:27
LABEL maintainer="admin@nfvpe.site"
RUN dnf install qpid-proton-c -y && dnf clean all
COPY --from=builder /tmp/metrics_consumer /
EXPOSE 8081
ENTRYPOINT ["/metrics_consumer"]
