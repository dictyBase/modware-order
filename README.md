# modware-order
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)  
![Continuous integration](https://github.com/dictyBase/modware-order/workflows/Continuous%20integration/badge.svg)
[![codecov](https://codecov.io/gh/dictyBase/modware-order/branch/develop/graph/badge.svg)](https://codecov.io/gh/dictyBase/modware-order)
[![Maintainability](https://api.codeclimate.com/v1/badges/f6b7067bd24eab14ba0d/maintainability)](https://codeclimate.com/github/dictyBase/modware-order/maintainability)   
![Last commit](https://badgen.net/github/last-commit/dictyBase/modware-order/develop)   
[![Funding](https://badgen.net/badge/Funding/Rex%20L%20Chisholm,dictyBase,DCR/yellow?list=|)](https://projectreporter.nih.gov/project_info_description.cfm?aid=10024726&icde=0)

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

# Misc badges
![Issues](https://badgen.net/github/issues/dictyBase/modware-order)
![Open Issues](https://badgen.net/github/open-issues/dictyBase/modware-order)
![Closed Issues](https://badgen.net/github/closed-issues/dictyBase/modware-order)  
![Total PRS](https://badgen.net/github/prs/dictyBase/modware-order)
![Open PRS](https://badgen.net/github/open-prs/dictyBase/modware-order)
![Closed PRS](https://badgen.net/github/closed-prs/dictyBase/modware-order)
![Merged PRS](https://badgen.net/github/merged-prs/dictyBase/modware-order)  
![Commits](https://badgen.net/github/commits/dictyBase/modware-order/develop)
![Branches](https://badgen.net/github/branches/dictyBase/modware-order)
![Tags](https://badgen.net/github/tags/dictyBase/modware-order)  
![GitHub repo size](https://img.shields.io/github/repo-size/dictyBase/modware-order?style=plastic)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/dictyBase/modware-order?style=plastic)
[![Lines of Code](https://badgen.net/codeclimate/loc/dictyBase/modware-order)](https://codeclimate.com/github/dictyBase/modware-order/code)  

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="http://cybersiddhu.github.com/"><img src="https://avatars3.githubusercontent.com/u/48740?v=4" width="100px;" alt=""/><br /><sub><b>Siddhartha Basu</b></sub></a><br /><a href="https://github.com/dictyBase/modware-order/issues?q=author%3Acybersiddhu" title="Bug reports">🐛</a> <a href="https://github.com/dictyBase/modware-order/commits?author=cybersiddhu" title="Code">💻</a> <a href="#content-cybersiddhu" title="Content">🖋</a> <a href="https://github.com/dictyBase/modware-order/commits?author=cybersiddhu" title="Documentation">📖</a> <a href="#maintenance-cybersiddhu" title="Maintenance">🚧</a></td>
    <td align="center"><a href="http://www.erichartline.net/"><img src="https://avatars3.githubusercontent.com/u/13489381?v=4" width="100px;" alt=""/><br /><sub><b>Eric Hartline</b></sub></a><br /><a href="https://github.com/dictyBase/modware-order/issues?q=author%3Awildlifehexagon" title="Bug reports">🐛</a> <a href="https://github.com/dictyBase/modware-order/commits?author=wildlifehexagon" title="Code">💻</a> <a href="#content-wildlifehexagon" title="Content">🖋</a> <a href="https://github.com/dictyBase/modware-order/commits?author=wildlifehexagon" title="Documentation">📖</a> <a href="#maintenance-wildlifehexagon" title="Maintenance">🚧</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!