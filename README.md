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

## Subcommand

```
NAME:
   modware-order start-server - starts the modware-order microservice with grpc backends

USAGE:
   modware-order start-server [command options] [arguments...]

OPTIONS:
   --arangodb-pass value, --pass value    arangodb database password [$ARANGODB_PASS]
   --arangodb-database value, --db value  arangodb database name [$ARANGODB_DATABASE]
   --arangodb-user value, --user value    arangodb database user [$ARANGODB_USER]
   --arangodb-host value, --host value    arangodb database host (default: "arangodb") [$ARANGODB_SERVICE_HOST]
   --arangodb-port value                  arangodb database port (default: "8529") [$ARANGODB_SERVICE_PORT]
   --is-secure                            flag for secured or unsecured arangodb endpoint
   --nats-host value                      nats messaging server host [$NATS_SERVICE_HOST]
   --nats-port value                      nats messaging server port [$NATS_SERVICE_PORT]
   --port value                           tcp port at which the server will be available (default: "9560")
   --order-collection value               arangodb collection for storing stock orders (default: "stock_order")
   --reflection, --ref                    flag for enabling server reflection
```

# API

### gRPC

The protocol buffer definitions and service apis are documented
[here](https://github.com/dictyBase/dictybaseapis/blob/master/dictybase/order/order.proto).
