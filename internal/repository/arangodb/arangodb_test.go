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
)

var ahost, aport, auser, apass, adb string
var coll driver.Collection
var collection = "stock_order"

func TestMain(m *testing.M) {
	adocker, err := aphdocker.NewArangoDockerWithImage("arangodb:3.3.19")
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
	p, _ := strconv.Atoi(aport)

	connP := &manager.ConnectParams{
		User:     auser,
		Pass:     apass,
		Database: adb,
		Host:     ahost,
		Port:     p,
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
