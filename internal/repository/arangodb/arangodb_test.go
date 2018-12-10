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

var gta *testarango.TestArango
var collection = "stock_orders"

func getConnectParams() *manager.ConnectParams {
	return &manager.ConnectParams{
		User:     gta.User,
		Pass:     gta.Pass,
		Database: gta.Database,
		Host:     gta.Host,
		Port:     gta.Port,
		Istls:    false,
	}
}

func newTestOrder() *order.NewOrder {
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
				Consumer:         "art@vandelayindustries.com",
				Payer:            "dr.van.nostrand@gmail.com",
				Purchaser:        "dr.van.nostrand@gmail.com",
				Items:            []string{"DBS2109858", "DBP8349822"},
			},
		},
	}
}

func TestMain(m *testing.M) {
	ta, err := testarango.NewTestArangoFromEnv(true)
	if err != nil {
		log.Fatalf("unable to construct new TestArango instance %s", err)
	}
	gta = ta
	dbh, err := ta.DB(ta.Database)
	if err != nil {
		log.Fatalf("unable to get database %s", err)
	}
	_, err = dbh.CreateCollection(collection, &driver.CreateCollectionOptions{})
	if err != nil {
		dbh.Drop()
		log.Fatalf("unable to create collection %s %s", collection, err)
	}
	code := m.Run()
	dbh.Drop()
	os.Exit(code)
}

func TestAddOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	no := newTestOrder()
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
	// assert.Equal(m.Status, no.Data.Attributes.Status, "should match the status")
	assert.Equal(m.Consumer, no.Data.Attributes.Consumer, "should match the consumer")
	assert.Equal(m.Payer, no.Data.Attributes.Payer, "should match the payer")
	assert.Equal(m.Purchaser, no.Data.Attributes.Purchaser, "should match the purchaser")
	assert.Equal(m.Items, no.Data.Attributes.Items, "should match the items")
}

func TestGetOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	no := newTestOrder()
	// add new test order
	m, err := repo.AddOrder(no)
	if err != nil {
		t.Fatalf("error in adding order %s", err)
	}
	// get test order by the key/ID of added test order
	g, err := repo.GetOrder(m.Key)
	if err != nil {
		t.Fatalf("error in getting order %s", err)
	}
	assert := assert.New(t)
	assert.Equal(g.Courier, no.Data.Attributes.Courier, "should match the courier")
	assert.Equal(g.CourierAccount, no.Data.Attributes.CourierAccount, "should match the courier account")
	assert.Equal(g.Comments, no.Data.Attributes.Comments, "should match the comments")
	assert.Equal(g.Payment, no.Data.Attributes.Payment, "should match the payment")
	assert.Equal(g.PurchaseOrderNum, no.Data.Attributes.PurchaseOrderNum, "should match the purchase order number")
	// assert.Equal(g.Status, no.Data.Attributes.Status, "should match the status")
	assert.Equal(g.Consumer, no.Data.Attributes.Consumer, "should match the consumer")
	assert.Equal(g.Payer, no.Data.Attributes.Payer, "should match the payer")
	assert.Equal(g.Purchaser, no.Data.Attributes.Purchaser, "should match the purchaser")
	assert.Equal(g.Items, no.Data.Attributes.Items, "should match the items")
}

func TestEditOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	no := newTestOrder()
	// add new test order
	m, err := repo.AddOrder(no)
	if err != nil {
		t.Fatalf("error in adding order %s", err)
	}
	// set the content to update
	testData := &order.OrderUpdate{
		Data: &order.OrderUpdate_Data{
			Type: "order",
			Id:   m.Key,
			Attributes: &order.OrderUpdateAttributes{
				Courier:  "UPS",
				Comments: "This is an updated test comment",
				Status:   1, // "Growing"
			},
		},
	}
	// edit test order by the key/ID of added test order
	e, err := repo.EditOrder(testData)
	if err != nil {
		t.Fatalf("error in editing order: %s", err)
	}
	assert := assert.New(t)
	// tests to make sure updated data matches passed in data
	assert.Equal(e.Courier, testData.Data.Attributes.Courier, "should match the new courier")
	assert.Equal(e.Comments, testData.Data.Attributes.Comments, "should match the new comments")
	// assert.Equal(e.Status, testData.Data.Attributes.Status, "should match the new status")

	// get the recently modified order so we can compare
	g, err := repo.GetOrder(m.Key)
	if err != nil {
		t.Fatalf("error in getting order: %s", err)
	}
	// make sure existing data wasn't overwritten by update
	assert.Equal(e.CourierAccount, g.CourierAccount, "should match the already existing courier account")
	assert.Equal(e.Courier, g.Courier, "should match the new courier")
}

// func TestListOrders(t *testing.T) {

// }
