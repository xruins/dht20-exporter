# syntax=docker/dockerfile:1@sha256:87999aa3d42bdc6bea60565083ee17e86d1f3339802f543c0d03998580f9cb89

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm@sha256:a9c020ee3d1508c7be5435c262434e3d3fc1d0e76a11afeb9ddae7d60bc86aa4 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN set -eux; \
  GOOS=${TARGETOS:-linux}; \
  GOARCH=${TARGETARCH:-amd64}; \
  unset GOARM || true; \
  if [ "$GOARCH" = "arm" ]; then \
    case "${TARGETVARIANT:-v7}" in \
      v5) GOARM=5 ;; \
      v6) GOARM=6 ;; \
      v7|"") GOARM=7 ;; \
      *) echo "unsupported TARGETVARIANT=${TARGETVARIANT} for GOARCH=arm" >&2; exit 1 ;; \
    esac; \
    export GOARM; \
  fi; \
  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -trimpath -ldflags="-s -w" -o /out/dht20-exporter .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:b7bb25d9f7c31d2bdd1982feb4dafcaf137703c7075dbe2febb41c24212b946f

COPY --from=build /out/dht20-exporter /dht20-exporter

EXPOSE 2112
ENTRYPOINT ["/dht20-exporter"]
