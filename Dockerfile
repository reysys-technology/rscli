# Build stage
FROM golang:1.25 AS builder

WORKDIR /build

ARG VERSION=dev

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY pkg pkg

# Build static binary with same flags as .goreleaser.yml
RUN CGO_ENABLED=0 go build \
    -tags 'netgo osusergo' \
    -ldflags="-s -w -extldflags '-static' -X main.version=${VERSION}" \
    -trimpath \
    -o rscli \
    ./cmd/rscli

FROM scratch

COPY --from=builder /build/rscli /rscli

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/rscli"]