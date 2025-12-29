# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS build

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

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/dht20-exporter /dht20-exporter

EXPOSE 2112
ENTRYPOINT ["/dht20-exporter"]
