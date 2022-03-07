FROM golang:1.16-alpine AS base

RUN set -eux \
    && apk update \
    && apk --no-cache add ca-certificates \
    && apk --no-cache add --virtual build-dependencies cmake g++ make unzip curl upx git

WORKDIR ${GOPATH}/src/github.com/yahoojapan/authorization-proxy

COPY go.mod .
COPY go.sum .

RUN GO111MODULE=on go mod download

FROM base AS builder

ENV APP_NAME authorization-proxy
ARG APP_VERSION='development version'

COPY . .

RUN BUILD_TIME=$(date -u +%Y%m%d-%H%M%S) \
    && GO_VERSION=$(go version | cut -d" " -f3,4) \
    && CGO_ENABLED=1 \
    CGO_CXXFLAGS="-g -Ofast -march=native" \
    CGO_FFLAGS="-g -Ofast -march=native" \
    CGO_LDFLAGS="-g -Ofast -march=native" \
    GOOS=$(go env GOOS) \
    GOARCH=$(go env GOARCH) \
    GO111MODULE=on \
    go build --ldflags "-s -w -linkmode 'external' -extldflags '-static -fPIC -m64 -pthread -std=c++11 -lstdc++' -X 'main.Version=${APP_VERSION} at ${BUILD_TIME} by ${GO_VERSION}'" -a -tags "cgo netgo" -installsuffix "cgo netgo" -o "${APP_NAME}" \
    && upx --best -o "/usr/bin/${APP_NAME}" "${APP_NAME}"

RUN apk del build-dependencies --purge \
    && rm -rf "${GOPATH}"

# Start From Scratch For Running Environment
FROM scratch
# FROM alpine:latest
LABEL maintainer "kpango <i.can.feel.gravity@gmail.com>"

ENV APP_NAME authorization-proxy

# Copy certificates for SSL/TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy permissions
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /usr/bin/${APP_NAME} /go/bin/${APP_NAME}

ENTRYPOINT ["/go/bin/authorization-proxy"]
