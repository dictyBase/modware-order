package arangodb

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

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

func newTestOrder(consumer string) *order.NewOrder {
	return &order.NewOrder{
		Data: &order.NewOrder_Data{
			Type: "order",
			Attributes: &order.NewOrderAttributes{
				Courier:          "FedEx",
				CourierAccount:   "9912378999",
				Comments:         "This is a test comment",
				Payment:          "Credit card",
				PurchaseOrderNum: "38975932199",
				Status:           order.OrderStatus_In_preparation,
				Consumer:         consumer,
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
	defer repo.ClearOrders()
	no := newTestOrder("art@vandelayindustries.com")
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
	assert.Equal(m.Status, no.Data.Attributes.Status.String(), "should match the status")
	assert.Equal(m.Consumer, no.Data.Attributes.Consumer, "should match the consumer")
	assert.Equal(m.Payer, no.Data.Attributes.Payer, "should match the payer")
	assert.Equal(m.Purchaser, no.Data.Attributes.Purchaser, "should match the purchaser")
	assert.Equal(m.Items, no.Data.Attributes.Items, "should match the items")
	assert.NotEmpty(m.Key, "should not have empty key/id")
}

func TestGetOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	defer repo.ClearOrders()
	no := newTestOrder("art@vandelayindustries.com")
	// add new test order
	m, err := repo.AddOrder(no)
	if err != nil {
		t.Fatalf("error in adding order %s", err)
	}
	// get test order by the key/ID of added test order
	g, err := repo.GetOrder(m.Key)
	if err != nil {
		t.Fatalf("error in getting order %s with ID %s", m.Key, err)
	}
	assert := assert.New(t)
	assert.Equal(g.Courier, no.Data.Attributes.Courier, "should match the courier")
	assert.Equal(g.CourierAccount, no.Data.Attributes.CourierAccount, "should match the courier account")
	assert.Equal(g.Comments, no.Data.Attributes.Comments, "should match the comments")
	assert.Equal(g.Payment, no.Data.Attributes.Payment, "should match the payment")
	assert.Equal(g.PurchaseOrderNum, no.Data.Attributes.PurchaseOrderNum, "should match the purchase order number")
	assert.Equal(g.Status, no.Data.Attributes.Status.String(), "should match the status")
	assert.Equal(g.Consumer, no.Data.Attributes.Consumer, "should match the consumer")
	assert.Equal(g.Payer, no.Data.Attributes.Payer, "should match the payer")
	assert.Equal(g.Purchaser, no.Data.Attributes.Purchaser, "should match the purchaser")
	assert.Equal(g.Items, no.Data.Attributes.Items, "should match the items")
	assert.Equal(len(g.Items), 2, "should match length of two items")
	assert.NotEmpty(g.Key, "should not have empty key/id")
	assert.True(m.CreatedAt.Equal(g.CreatedAt), "should match created time of order")
	assert.True(m.UpdatedAt.Equal(g.UpdatedAt), "should match updated time of order")

	ne, err := repo.GetOrder("1")
	if err != nil {
		t.Fatalf(
			"error in fetching order %s with ID %s",
			"1",
			err,
		)
	}
	assert.True(ne.NotFound, "entry should not exist")
}

func TestEditOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	defer repo.ClearOrders()
	no := newTestOrder("art@vandelayindustries.com")
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
				Status:   order.OrderStatus_Growing,
			},
		},
	}
	// edit test order by providing updated data
	e, err := repo.EditOrder(testData)
	if err != nil {
		t.Fatalf("error in editing order %s with ID %s", m.Key, err)
	}
	assert := assert.New(t)
	// tests to make sure updated data matches passed in data
	assert.Equal(e.Courier, testData.Data.Attributes.Courier, "should match the new courier")
	assert.Equal(e.Comments, testData.Data.Attributes.Comments, "should match the new comments")
	assert.Equal(e.Status, testData.Data.Attributes.Status.String(), "should match the new status")

	// get the recently modified order so we can compare
	g, err := repo.GetOrder(m.Key)
	if err != nil {
		t.Fatalf("error in getting order %s with ID %s", m.Key, err)
	}
	// make sure existing data wasn't overwritten by update
	assert.Equal(g.CourierAccount, m.CourierAccount, "should match the already existing courier account")
	assert.Equal(e.Courier, g.Courier, "should match the new courier")
	assert.NotEqual(g.Courier, m.Courier, "should not match the already existing courier")
	assert.NotEqual(g.Status, m.Status, "should not match the already existing status")

	// set data with nonexistent ID
	ed := &order.OrderUpdate{
		Data: &order.OrderUpdate_Data{
			Type: "order",
			Id:   "1",
			Attributes: &order.OrderUpdateAttributes{
				Comments: "This is an updated test comment",
			},
		},
	}
	ee, err := repo.EditOrder(ed)
	if err != nil {
		t.Fatalf("error in editing order: %s", err)
	}
	assert.True(ee.NotFound, "entry should not exist")
}

func TestListOrders(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	defer repo.ClearOrders()
	// add 15 new test orders
	for i := 1; i <= 15; i++ {
		no := newTestOrder(fmt.Sprintf("%s@kramericaindustries.com", RandString(10)))
		_, err := repo.AddOrder(no)
		if err != nil {
			t.Fatalf("error in adding order %s", err)
		}
	}
	// get first five results
	lo, err := repo.ListOrders(0, 4)
	if err != nil {
		t.Fatalf("error in getting first five orders %s", err)
	}
	assert := assert.New(t)
	assert.Equal(len(lo), 5, "should match the provided limit number + 1")

	for _, order := range lo {
		assert.Equal(order.Courier, "FedEx", "should match the courier")
		assert.NotEmpty(order.Key, "should not have empty key/id")
	}
	// compare timestamps for first two results
	if lo[1].CreatedAt.UnixNano() > lo[0].CreatedAt.UnixNano() {
		t.Fatalf("the created_at date of the second item should be older than the first item")
	}
	assert.NotEqual(lo[0].Consumer, lo[1].Consumer, "should have different consumers")
	// convert fifth result to numeric timestamp in milliseconds
	// so we can use this as cursor
	ti := lo[4].CreatedAt.UnixNano() / 1000000

	// get next five results (5-9)
	lo2, err := repo.ListOrders(ti, 4)
	if err != nil {
		t.Fatalf("error in getting orders 5-9 %s", err)
	}
	assert.Equal(len(lo2), 5, "should match the provided limit number + 1")
	for _, order := range lo2 {
		assert.Equal(order.Courier, "FedEx", "should match the courier")
		assert.NotEmpty(order.Key, "should not have empty key/id")
	}
	// compare timestamps for first two results
	if lo2[1].CreatedAt.UnixNano() > lo2[0].CreatedAt.UnixNano() {
		t.Fatalf("the created_at date of the second item should be older than the first item")
	}
	assert.NotEqual(lo2[0].Consumer, lo2[1].Consumer, "should have different consumers")

	// convert ninth result to numeric timestamp
	ti2 := lo2[4].CreatedAt.UnixNano() / 1000000
	// get last five results (9-13)
	lo3, err := repo.ListOrders(ti2, 4)
	if err != nil {
		t.Fatalf("error in getting orders 9-13 %s", err)
	}
	assert.Equal(len(lo3), 5, "should match the provided limit number + 1")
	for _, order := range lo3 {
		assert.Equal(order.Courier, "FedEx", "should match the courier")
		assert.NotEmpty(order.Key, "should not have empty key/id")
	}
	// compare timestamps for first two results
	if lo3[1].CreatedAt.UnixNano() > lo3[0].CreatedAt.UnixNano() {
		t.Fatalf("the created_at date of the second item should be older than the first item")
	}

	// convert 13th result to numeric timestamp
	ti3 := lo3[4].CreatedAt.UnixNano() / 1000000
	// get last results
	lo4, err := repo.ListOrders(ti3, 4)
	if err != nil {
		t.Fatalf("error in getting orders 13-15 %s", err)
	}
	assert.Equal(len(lo4), 3, "should only bring last three results")
	// compare timestamps for first two results
	if lo4[1].CreatedAt.UnixNano() > lo4[0].CreatedAt.UnixNano() {
		t.Fatalf("the created_at date of the second item should be older than the first item")
	}
}

const (
	charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seedRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(
			b,
			charset[seedRand.Intn(len(charset))],
		)
	}
	return string(b)
}

func RandString(length int) string {
	return stringWithCharset(length, charSet)
}
