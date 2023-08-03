package main

import (
	"log"
	"os"

	"github.com/dictyBase/aphgrpc"
	arango "github.com/dictyBase/arangomanager/command/flag"
	"github.com/dictyBase/modware-order/internal/app/server"
	"github.com/dictyBase/modware-order/internal/app/validate"
	"github.com/urfave/cli"
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
			Before: validate.ServerArgs,
			Flags:  serverFlags(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("error in running command %s", err)
	}
}

func serverFlags() []cli.Flag {
	flg := make([]cli.Flag, 0)
	flg = append(flg, arango.ArangoFlags()...)
	flg = append(flg, aphgrpc.NatsFlag()...)
	flg = append(flg, []cli.Flag{
		cli.StringFlag{
			Name:   "arangodb-database, db",
			EnvVar: "ARANGODB_DATABASE",
			Usage:  "arangodb database name",
			Value:  "stock",
		},
		cli.StringFlag{
			Name:  "port",
			Usage: "tcp port at which the server will be available",
			Value: "9599",
		},
		cli.StringFlag{
			Name:  "order-collection",
			Usage: "arangodb collection for storing stock orders",
			Value: "stock_order",
		},
		cli.BoolTFlag{
			Name:  "reflection, ref",
			Usage: "flag for enabling server reflection",
		},
	}...)

	return flg
}
