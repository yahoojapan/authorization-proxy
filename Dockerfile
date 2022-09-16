FROM golang:1.18-alpine AS base

RUN set -eux \
    && apk --no-cache add ca-certificates \
    && apk --no-cache add --virtual build-dependencies cmake g++ make unzip curl git libcap

WORKDIR ${GOPATH}/src/github.com/yahoojapan/authorization-proxy

COPY go.mod .
COPY go.sum .

RUN GO111MODULE=on go mod download

FROM base AS builder

ENV APP_NAME authorization-proxy
ARG APP_VERSION='development version'

COPY . .

RUN adduser -H -S ${APP_NAME}

RUN BUILD_TIME=$(date -u +%Y%m%d-%H%M%S) \
    && GO_VERSION=$(go version | cut -d" " -f3,4) \
    && CGO_ENABLED=1 \
    CGO_CXXFLAGS="-g -Ofast -march=native" \
    CGO_FFLAGS="-g -Ofast -march=native" \
    CGO_LDFLAGS="-g -Ofast -march=native" \
    GOOS=$(go env GOOS) \
    GOARCH=$(go env GOARCH) \
    GO111MODULE=on \
    go build -ldflags "-X 'main.Version=${VERSION} at ${BUILD_TIME} by ${GO_VERSION}' -linkmode=external" -a -o "/usr/bin/${APP_NAME}"

# allow well-known port binding
RUN setcap 'cap_net_bind_service=+ep' "/usr/bin/${APP_NAME}"

# confirm dependency libraries & cleanup
RUN ldd "/usr/bin/${APP_NAME}"\
    && apk del build-dependencies --purge \
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
# Copy our dynamic-linked executable and library
COPY --from=builder /usr/bin/${APP_NAME} /go/bin/${APP_NAME}
COPY --from=builder /lib/ld-musl-x86_64.so* /lib/
# Copy user
COPY --from=builder /etc/passwd /etc/passwd
USER ${APP_NAME}

HEALTHCHECK NONE
ENTRYPOINT ["/go/bin/authorization-proxy"]
