FROM golang:1.11.5-alpine3.8
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
RUN apk add --no-cache git build-base
RUN mkdir -p /modware-order
WORKDIR /modware-order
COPY go.mod cmd/modware-order/main.go ./
ADD internal internal
RUN go get ./... && go mod tidy && go build -o app

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=0 /modware-order/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]