FROM golang:1.16.5-buster AS builder
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
ENV GOPROXY https://proxy.golang.org
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN apt-get -qq update \
    && apt-get -yqq install upx
RUN mkdir -p /modware-order
WORKDIR /modware-order
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD cmd cmd
ADD internal internal
RUN go build \
    -a \
    -ldflags "-s -w -extldflags '-static'" \
    -installsuffix cgo \
    -tags netgo \
    -o /bin/app \
    cmd/modware-order/main.go
RUN strip /bin/app \
    && upx -q -9 /bin/app

FROM gcr.io/distroless/static
COPY --from=builder /bin/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
