# build golang application
FROM golang:1.24.4 AS builder

ARG APP_NAME
ARG BUILD_PATH=./cmd/${APP_NAME}

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o ${APP_NAME} ${BUILD_PATH}

# build docker image
FROM alpine:3.22.0

ARG APP_NAME=prometheus-net-discovery
ARG UID=65532
ARG GID=65532

RUN apk add --no-cache \
      nmap \
      curl \
    && rm -rf /var/cache/apk/*

RUN addgroup -g ${GID} ${APP_NAME} && adduser -D -u ${UID} -G ${APP_NAME} ${APP_NAME}

USER ${APP_NAME}

WORKDIR /app

COPY --from=builder /build/${APP_NAME} /bin/

RUN ln -s /bin/${APP_NAME} /app/entrypoint

ENTRYPOINT ["/app/entrypoint"]
