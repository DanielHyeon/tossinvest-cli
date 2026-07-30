# syntax=docker/dockerfile:1.7
FROM golang:1.25.0-alpine3.22 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" -o /out/tossctl ./cmd/tossctl

FROM alpine:3.22.1

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 tossos \
    && adduser -S -D -H -u 10001 -G tossos tossos \
    && mkdir -p /var/lib/tossos/config /var/lib/tossos/data \
    && chown -R 10001:10001 /var/lib/tossos

COPY --from=build /out/tossctl /usr/local/bin/tossctl
COPY deploy/container-entrypoint.sh /usr/local/bin/tossos-entrypoint

ENV XDG_CONFIG_HOME=/var/lib/tossos/config \
    XDG_DATA_HOME=/var/lib/tossos/data \
    TOSSOS_CONTAINER=1

USER 10001:10001
EXPOSE 37085
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-check-certificate -qO- https://127.0.0.1:37085/healthz | grep -qx ok

ENTRYPOINT ["/usr/local/bin/tossos-entrypoint"]
CMD ["console", "--port", "37085"]
