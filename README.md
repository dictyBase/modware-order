# modware-order

[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)  
[![Go Report Card](https://goreportcard.com/badge/github.com/dictybase/modware-order)](https://goreportcard.com/report/github.com/dictybase/modware-order)
![](https://github.com/dictyBase/modware-order/workflows/.github/workflows/ci.yml/badge.svg)
[![godoc](https://godoc.org/github.com/dictyBase/modware-order?status.svg)](https://godoc.org/github.com/dictyBase/modware-order)
[![codecov](https://codecov.io/gh/dictyBase/modware-order/branch/develop/graph/badge.svg)](https://codecov.io/gh/dictyBase/modware-order)  
[![Technical debt](https://badgen.net/codeclimate/tech-debt/dictyBase/modware-order)](https://codeclimate.com/github/dictyBase/modware-order/trends/technical_debt)
[![Issues](https://badgen.net/codeclimate/issues/dictyBase/modware-order)](https://codeclimate.com/github/dictyBase/modware-order/issues)
[![Maintainability percentage](https://badgen.net/codeclimate/maintainability-percentage/dictyBase/modware-order)](https://codeclimate.com/github/dictyBase/modware-order)  
![Issues](https://badgen.net/github/issues/dictyBase/modware-order)
![Open Issues](https://badgen.net/github/open-issues/dictyBase/modware-order)
![Closed Issues](https://badgen.net/github/closed-issues/dictyBase/modware-order)  
![Total PRS](https://badgen.net/github/prs/dictyBase/modware-order)
![Open PRS](https://badgen.net/github/open-prs/dictyBase/modware-order)
![Closed PRS](https://badgen.net/github/closed-prs/dictyBase/modware-order)
![Merged PRS](https://badgen.net/github/merged-prs/dictyBase/modware-order)  
![Commits](https://badgen.net/github/commits/dictyBase/modware-order/develop)
![Last commit](https://badgen.net/github/last-commit/dictyBase/modware-order/develop)
![Branches](https://badgen.net/github/branches/dictyBase/modware-order)
![Tags](https://badgen.net/github/tags/dictyBase/modware-order)  
![GitHub repo size](https://img.shields.io/github/repo-size/dictyBase/modware-order?style=plastic)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/dictyBase/modware-order?style=plastic)
[![Lines of Code](https://badgen.net/codeclimate/loc/dictyBase/modware-order)](https://codeclimate.com/github/dictyBase/modware-order/code)  
[![Funding](https://badgen.net/badge/NIGMS/Rex%20L%20Chisholm,dictyBase/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=9476993)
[![Funding](https://badgen.net/badge/NIGMS/Rex%20L%20Chisholm,DSC/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=9438930)

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

## Default Names

There is only one collection, and its default name is **stock_order**.

# API

### gRPC

The protocol buffer definitions and service apis are documented
[here](https://github.com/dictyBase/dictybaseapis/blob/master/dictybase/order/order.proto).
