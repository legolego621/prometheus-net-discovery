# build golang application
FROM golang:1.24.2 AS builder

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o prometheus-net-discovery /build/cmd/prometheus-net-discovery

# build docker image
FROM alpine:3.22.0

ARG NAME=prometheus-net-discovery
ARG VERSION=latest
ARG UID=65532
ARG GID=65532

RUN apk add --no-cache \
      nmap \
      curl \
    && rm -rf /var/cache/apk/*

RUN addgroup -g ${GID} ${NAME} && adduser -D -u ${UID} -G ${NAME} ${NAME}

ENV HOME=/home/${NAME}

USER ${NAME}

WORKDIR /app

COPY --from=builder /build/${NAME} /bin/

ENTRYPOINT ["/bin/prometheus-net-discovery"]
