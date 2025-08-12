ARG GOLANG_VERSION=1.24.6
ARG TARGETARCH=amd64
ARG VERSION=latest
ARG BASE_REGISTRY=registry.access.redhat.com
ARG BASE_IMAGE=ubi8-micro
ARG PLATFORM=amd64

FROM docker.io/library/golang:${GOLANG_VERSION}-alpine AS builder
LABEL stage=featureservbuilder

# Install build dependencies for CGO and C++
RUN apk add --no-cache gcc g++ musl-dev libstdc++-dev

ARG TARGETARCH
ARG VERSION

WORKDIR /app
COPY . ./

# Native build on target platform
RUN go build -v -ldflags "-s -w -X github.com/tobilg/duckdb_featureserv/internal/conf.setVersion=${VERSION}"

FROM --platform=${TARGETARCH} ${BASE_REGISTRY}/${BASE_IMAGE} AS multi-stage

COPY --from=builder /app/duckdb_featureserv .
COPY --from=builder /app/assets ./assets

VOLUME ["/config"]
VOLUME ["/assets"]

USER 1001
EXPOSE 9000

ENTRYPOINT ["./duckdb_featureserv"]
CMD []

FROM --platform=${PLATFORM} ${BASE_REGISTRY}/${BASE_IMAGE} AS local

ADD ./duckdb_featureserv .
ADD ./assets ./assets

VOLUME ["/config"]
VOLUME ["/assets"]

USER 1001
EXPOSE 9000

ENTRYPOINT ["./duckdb_featureserv"]
CMD []

# To build
# make APPVERSION=1.1 clean build docker

# To build using binaries from golang docker image
# make APPVERSION=1.1 clean multi-stage-docker
