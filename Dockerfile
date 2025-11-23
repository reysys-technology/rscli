FROM alpine:latest AS certs

WORKDIR /build


FROM scratch

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/rscli /usr/bin/
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/rscli"]