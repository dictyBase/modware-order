package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/dictyBase/apihelpers/aphgrpc"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/app/service"
	"github.com/dictyBase/modware-order/internal/message/nats"
	"github.com/dictyBase/modware-order/internal/repository/arangodb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	gnats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// RunServer starts and runs the server.
func RunServer(clt *cli.Context) error {
	arPort, _ := strconv.Atoi(clt.String("arangodb-port"))
	connP := &manager.ConnectParams{
		User:     clt.String("arangodb-user"),
		Pass:     clt.String("arangodb-pass"),
		Database: clt.String("arangodb-database"),
		Host:     clt.String("arangodb-host"),
		Port:     arPort,
		Istls:    clt.Bool("is-secure"),
	}
	anrepo, err := arangodb.NewOrderRepo(connP, "stock_order")
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("cannot connect to arangodb order repository %s",
				err.Error(),
			), 2)
	}
	msn, err := nats.NewPublisher(
		clt.String("nats-host"),
		clt.String("nats-port"),
		gnats.MaxReconnects(-1),
		gnats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("cannot connect to messaging server %s",
				err.Error(),
			), 2)
	}
	grpcS := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(getLogger(clt)),
		),
	)
	order.RegisterOrderServiceServer(
		grpcS, service.NewOrderService(anrepo, msn,
			aphgrpc.TopicsOption(map[string]string{
				"orderCreate": "OrderService.Create",
				"orderUpdate": "OrderService.Update",
			}),
		),
	)
	if clt.Bool("reflection") {
		// register reflection service on gRPC server
		reflection.Register(grpcS)
	}
	// create listener
	endP := fmt.Sprintf(":%s", clt.String("port"))
	lis, err := net.Listen("tcp", endP)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("failed to listen %s", err), 2)
	}
	log.Printf("starting grpc server on %s", endP)
	if err := grpcS.Serve(lis); err != nil {
		return fmt.Errorf("error in starting grpc server %s", err)
	}

	return nil
}

func getLogger(clt *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch clt.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := clt.GlobalString("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}

	return logrus.NewEntry(log)
}
