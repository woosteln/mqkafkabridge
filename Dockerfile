# build stage
FROM golang:alpine AS build-env
# Install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh && \
    apk add build-base

# Build librdkafka
RUN mkdir /build && cd /build && git clone -b v0.11.4 https://github.com/edenhill/librdkafka.git librdkafka \
  && cd librdkafka \
  && ./configure --prefix=/usr \
  && make \
  && make install

# Test and build muapi
ADD . /go/src/github.com/woosteln/mqkafkabridge
RUN go get -u github.com/golang/dep/cmd/dep
RUN cd /go/src/github.com/woosteln/mqkafkabridge && \
  dep ensure && \
  CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/mqkafkabridge

# final stage
FROM alpine:3.7
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=build-env /go/src/github.com/woosteln/mqkafkabridge/bin/mqkafkabridge /usr/bin/mqkafkabridge
COPY --from=build-env /usr/lib/librdkafka* /usr/local/lib/
COPY --from=build-env /usr/lib/pkgconfig/rdkafka* /usr/lib/pkgconfig/
COPY --from=build-env /usr/include/librdkafka/* /usr/include/librdkafka/

RUN chmod +x /usr/bin/mqkafkabridge
ENTRYPOINT ["/usr/bin/mqkafkabridge"]
CMD ["/usr/bin/mqkafkabridge"]