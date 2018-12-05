package main

import (
	"os"

	"github.com/dictyBase/modware-order/internal/app/server"
	"github.com/dictyBase/modware-order/internal/app/validate"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "modware-order"
	app.Usage = "cli for modware-order microservice"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-format",
			Usage: "format of the logging out, either of json or text.",
			Value: "json",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "log level for the application",
			Value: "error",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "start-server",
			Usage:  "starts the modware-order microservice with grpc backends",
			Action: server.RunServer,
			Before: validate.ValidateServerArgs,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "arangodb-pass, pass",
					EnvVar: "ARANGODB_PASS",
					Usage:  "arangodb database password",
				},
				cli.StringFlag{
					Name:   "arangodb-database, db",
					EnvVar: "ARANGODB_DATABASE",
					Usage:  "arangodb database name",
				},
				cli.StringFlag{
					Name:   "arangodb-user, user",
					EnvVar: "ARANGODB_USER",
					Usage:  "arangodb database user",
				},
				cli.StringFlag{
					Name:   "arangodb-host, host",
					Value:  "arangodb",
					EnvVar: "ARANGODB_SERVICE_HOST",
					Usage:  "arangodb database host",
				},
				cli.StringFlag{
					Name:   "arangodb-port",
					EnvVar: "ARANGODB_SERVICE_PORT",
					Usage:  "arangodb database port",
					Value:  "8529",
				},
				cli.BoolTFlag{
					Name:  "is-secure",
					Usage: "flag for secured or unsecured arangodb endpoint",
				},
				cli.StringFlag{
					Name:   "nats-host",
					EnvVar: "NATS_SERVICE_HOST",
					Usage:  "nats messaging server host",
				},
				cli.StringFlag{
					Name:   "nats-port",
					EnvVar: "NATS_SERVICE_PORT",
					Usage:  "nats messaging server port",
				},
				cli.StringFlag{
					Name:  "port",
					Usage: "tcp port at which the server will be available",
					Value: "9560",
				},
				cli.StringFlag{
					Name:  "order-collection",
					Usage: "arangodb collection for storing stock orders",
					Value: "stock_order",
				},
			},
		},
	}
	app.Run(os.Args)
}
