package arangodb

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/apihelpers/aphdocker"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

var ahost, aport, auser, apass, adb string
var coll driver.Collection
var collection = "stock_order"

func newOrder(id string) *order.NewOrderAttributes {
	return &order.NewOrderAttributes{
		Courier:  "FedEx",
		Comments: "This is a test comment",
		Payment:  "Credit card",
	}
}

func TestMain(m *testing.M) {
	if len(os.Getenv("DOCKER_HOST")) == 0 {
		os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
	}
	if len(os.Getenv("DOCKER_API_VERSION")) == 0 {
		os.Setenv("DOCKER_API_VERSION", "1.35")
	}

	adocker, err := aphdocker.NewArangoDockerWithImage("arangodb:3.3.19")
	adocker.Debug = true
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	aresource, err := adocker.Run()
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	client, err := adocker.RetryConnection()
	if err != nil {
		log.Fatalf("unable to get client connection %s", err)
	}
	adb = aphdocker.RandString(6)
	dbh, err := client.CreateDatabase(context.Background(), adb, &driver.CreateDatabaseOptions{})
	if err != nil {
		log.Fatalf("could not create arangodb database %s %s\n", adb, err)
	}

	coll, err = dbh.CreateCollection(context.Background(), collection, &driver.CreateCollectionOptions{})
	if err != nil {
		log.Fatalf("could not create arangodb collection %s", collection)
	}
	auser = adocker.GetUser()
	apass = adocker.GetPassword()
	ahost = adocker.GetIP()
	aport = adocker.GetPort()
	code := m.Run()
	if err = adocker.Purge(aresource); err != nil {
		log.Fatalf("unable to remove arangodb container %s\n", err)
	}
	os.Exit(code)
}

func TestGetOrder(t *testing.T) {
	// convert port string to int
	port, _ := strconv.Atoi(aport)

	connP := &manager.ConnectParams{
		User:     auser,
		Pass:     apass,
		Database: adb,
		Host:     ahost,
		Port:     port,
		Istls:    false,
	}
	_, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to data source %s", err)
	}
}

// func TestAddOrder(t *testing.T) {

// }

// func TestEditOrder(t *testing.T) {

// }

// func TestListOrders(t *testing.T) {

// }
