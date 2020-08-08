FROM golang:1.14.6-stretch

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ARG GO_PROXY
ENV GOPROXY=${GO_PROXY}

COPY go.sum go.mod ./

WORKDIR ./

COPY . .

RUN make go_deps && make build

FROM alpine:3.12.0

RUN apk add --update --no-cache \
      tzdata \
      ca-certificates \
      bash \
    && \

COPY --from=0 /bin/sfu-load-test /usr/local/bin/sfu-load-test

ENTRYPOINT ["/usr/local/bin/sfu-load-test"]
#CMD ["-c", "/configs/sfu.toml"]