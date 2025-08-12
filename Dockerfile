ARG GOLANG_VERSION
ARG TARGETARCH
ARG VERSION
ARG BASE_REGISTRY
ARG BASE_IMAGE
ARG PLATFORM

FROM --platform=${PLATFORM} docker.io/library/golang:${GOLANG_VERSION}-alpine AS builder
LABEL stage=featureservbuilder

ARG TARGETARCH
ARG VERSION

WORKDIR /app
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -v -ldflags "-s -w -X main.programVersion=${VERSION}"

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
