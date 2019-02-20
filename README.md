# modware-order

[![GoDoc](https://godoc.org/github.com/dictybase/modware-order?status.svg)](https://godoc.org/github.com/dictybase/modware-order)
[![Go Report Card](https://goreportcard.com/badge/github.com/dictybase/modware-order)](https://goreportcard.com/report/github.com/dictybase/modware-order)

dictyBase API server to manage order of biological stocks. The API server supports gRPC protocol for data exchange.

## Usage

```
NAME:
   modware-order - cli for modware-order microservice

USAGE:
   modware-order [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     start-server  starts the modware-order microservice with grpc backends
     help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-format value  format of the logging out, either of json or text. (default: "json")
   --log-level value   log level for the application (default: "error")
   --help, -h          show help
   --version, -v       print the version
```

# API

### gRPC

The protocol buffer definitions and service apis are documented
[here](https://github.com/dictyBase/dictybaseapis/blob/master/dictybase/order/order.proto).
