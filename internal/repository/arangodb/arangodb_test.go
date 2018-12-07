package arangodb

import (
	"log"
	"os"
	"testing"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/arangomanager/testarango"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/stretchr/testify/assert"
)

var ta *testarango.TestArango
var collection = "stock_orders"

func getConnectParams() *manager.ConnectParams {
	return &manager.ConnectParams{
		User:     ta.User,
		Pass:     ta.Pass,
		Database: ta.Database,
		Host:     ta.Host,
		Port:     ta.Port,
		Istls:    false,
	}
}

func newOrder() *order.NewOrder {
	return &order.NewOrder{
		Data: &order.NewOrder_Data{
			Type: "order",
			Attributes: &order.NewOrderAttributes{
				Courier:          "FedEx",
				CourierAccount:   "9912378999",
				Comments:         "This is a test comment",
				Payment:          "Credit card",
				PurchaseOrderNum: "38975932199",
				Status:           0, // "In_preparation"
			},
		},
	}
}

func TestMain(m *testing.M) {
	ta, err := testarango.NewTestArangoFromEnv(true)
	if err != nil {
		log.Fatal(err)
	}
	dbh, err := ta.DB(ta.Database)
	if err != nil {
		log.Fatal(err)
	}
	dbh.CreateCollection(collection, &driver.CreateCollectionOptions{})
	if err != nil {
		log.Fatalf("unable to create collection %s %s", collection, err)
	}
	// clean up the database at the end
	defer dbh.Drop()

	code := m.Run()
	os.Exit(code)
}

func TestAddOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	no := newOrder()
	m, err := repo.AddOrder(no)
	if err != nil {
		t.Fatalf("error in adding order %s", err)
	}
	assert := assert.New(t)
	assert.Equal(m.Courier, no.Data.Attributes.Courier, "should match the courier")
	assert.Equal(m.CourierAccount, no.Data.Attributes.CourierAccount, "should match the courier account")
	assert.Equal(m.Comments, no.Data.Attributes.Comments, "should match the comments")
	assert.Equal(m.Payment, no.Data.Attributes.Payment, "should match the payment")
	assert.Equal(m.PurchaseOrderNum, no.Data.Attributes.PurchaseOrderNum, "should match the purchase order number")
	assert.Equal(m.Status, no.Data.Attributes.Status, "should match the status")
}

// func TestGetOrder(t *testing.T) {
// 	connP := getConnectParams()
// 	_, err := NewOrderRepo(connP, collection)
// 	if err != nil {
// 		t.Fatalf("error in connecting to data source %s", err)
// 	}
// }

// func TestEditOrder(t *testing.T) {

// }

// func TestListOrders(t *testing.T) {

// }
