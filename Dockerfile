# syntax=docker/dockerfile:1@sha256:b6afd42430b15f2d2a4c5a02b919e98a525b785b1aaff16747d2f623364e39b6

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm@sha256:2f768d462dbffbb0f0b3a5171009f162945b086f326e0b2a8fd5d29c3219ff14 AS build

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

FROM gcr.io/distroless/static-debian12:nonroot@sha256:cba10d7abd3e203428e86f5b2d7fd5eb7d8987c387864ae4996cf97191b33764

COPY --from=build /out/dht20-exporter /dht20-exporter

EXPOSE 2112
ENTRYPOINT ["/dht20-exporter"]
