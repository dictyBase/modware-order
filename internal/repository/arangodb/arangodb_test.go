package arangodb

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/apihelpers/aphgrpc"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/arangomanager/query"
	"github.com/dictyBase/arangomanager/testarango"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/model"
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

func toTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1000000
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
	assert.Len(g.Items, 2, "should match length of two items")
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
				Courier:          "UPS",
				CourierAccount:   "99999999",
				Comments:         "This is an updated test comment",
				Payment:          "Check",
				PurchaseOrderNum: "33333333",
				Items:            []string{"xyz", "abc"},
				Status:           order.OrderStatus_Growing,
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
	assert.Equal(e.CourierAccount, testData.Data.Attributes.CourierAccount, "should match the new courier account")
	assert.Equal(e.Comments, testData.Data.Attributes.Comments, "should match the new comments")
	assert.Equal(e.Payment, testData.Data.Attributes.Payment, "should match the new payment")
	assert.Equal(e.PurchaseOrderNum, testData.Data.Attributes.PurchaseOrderNum, "should match the new purchase order number")
	assert.ElementsMatch(e.Items, testData.Data.Attributes.Items, "should match the new items")
	assert.Equal(e.Status, testData.Data.Attributes.Status.String(), "should match the new status")

	// get the recently modified order so we can compare
	g, err := repo.GetOrder(m.Key)
	if err != nil {
		t.Fatalf("error in getting order %s with ID %s", m.Key, err)
	}
	// make sure existing data wasn't overwritten by update
	assert.Equal(g.Payer, m.Payer, "should match the already existing payer")
	assert.Equal(e.Courier, g.Courier, "should match the new courier")

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
	lo, err := repo.ListOrders(&order.ListParameters{Limit: 4})
	if err != nil {
		t.Fatalf("error in getting first five orders %s", err)
	}
	assert := assert.New(t)
	assert.Len(lo, 5, "should match the provided limit number + 1")

	for _, order := range lo {
		assert.Equal(order.Courier, "FedEx", "should match the courier")
		assert.NotEmpty(order.Key, "should not have empty key/id")
	}
	assert.NotEqual(lo[0].Consumer, lo[1].Consumer, "should have different consumers")
	// convert fifth result to numeric timestamp in milliseconds
	// so we can use this as cursor
	ti := toTimestamp(lo[len(lo)-1].CreatedAt)

	// get next five results (5-9)
	lo2, err := repo.ListOrders(&order.ListParameters{Cursor: ti, Limit: 4})
	if err != nil {
		t.Fatalf("error in getting orders 5-9 %s", err)
	}
	assert.Len(lo2, 5, "should match the provided limit number + 1")
	assert.Exactly(lo2[0], lo[len(lo)-1], "last item from first five results and first item from next five results should be the same")
	assert.NotEqual(lo2[0].Consumer, lo2[1].Consumer, "should have different consumers")

	// convert ninth result to numeric timestamp
	ti2 := toTimestamp(lo2[len(lo2)-1].CreatedAt)
	// get last five results (9-13)
	lo3, err := repo.ListOrders(&order.ListParameters{Cursor: ti2, Limit: 4})
	if err != nil {
		t.Fatalf("error in getting orders 9-13 %s", err)
	}
	assert.Len(lo3, 5, "should match the provided limit number + 1")
	assert.Exactly(lo3[0], lo2[len(lo2)-1], "last item from previous five results and first item from next five results should be the same")

	// convert 13th result to numeric timestamp
	ti3 := toTimestamp(lo3[len(lo3)-1].CreatedAt)
	// get last results
	lo4, err := repo.ListOrders(&order.ListParameters{Cursor: ti3, Limit: 4})
	if err != nil {
		t.Fatalf("error in getting orders 13-15 %s", err)
	}
	assert.Len(lo4, 3, "should only bring last three results")
	assert.Exactly(lo3[4], lo4[0], "last item from previous five results and first item from next three results should be the same")
	testModelListSort(lo, t)
	testModelListSort(lo2, t)
	testModelListSort(lo3, t)
	testModelListSort(lo4, t)

	sf, err := repo.ListOrders(&order.ListParameters{Limit: 100, Filter: convertFilterToQuery("courier===FedEx")})
	if err != nil {
		t.Fatalf("error in getting orders with courier filter %s", err)
	}
	assert.Len(sf, 15, "should list all 15 orders")

	sc, err := repo.ListOrders(&order.ListParameters{Cursor: toTimestamp(sf[5].CreatedAt), Limit: 100, Filter: convertFilterToQuery("courier===FedEx")})
	if err != nil {
		t.Fatalf("error in getting orders with cursor and courier filter %s", err)
	}
	assert.Len(sc, 10, "should list last 10 orders")

	sn, err := repo.ListOrders(&order.ListParameters{Cursor: toTimestamp(sf[5].CreatedAt), Limit: 100, Filter: convertFilterToQuery("courier===UPS")})
	if err != nil {
		t.Fatalf("error in getting orders with cursor and courier filter %s", err)
	}
	assert.Len(sn, 0, "should list last no UPS orders")
}

func TestLoadOrder(t *testing.T) {
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	if err != nil {
		t.Fatalf("error in connecting to order repository %s", err)
	}
	defer repo.ClearOrders()
	tm, _ := time.Parse("2006-01-02 15:04:05", "2010-03-30 14:40:58")
	eo := &order.ExistingOrder{
		Data: &order.ExistingOrder_Data{
			Type: "order",
			Attributes: &order.ExistingOrderAttributes{
				CreatedAt: aphgrpc.TimestampProto(tm),
				Consumer:  "super@c.org",
				Payer:     "super@c.org",
				Purchaser: "super@c.org",
				Items:     []string{"DBS2109858", "DBP8349822"},
			},
		},
	}
	m, err := repo.LoadOrder(eo)
	if err != nil {
		t.Fatalf("error in loading order %s", err)
	}
	assert := assert.New(t)
	assert.True(m.CreatedAt.Equal(tm), "should match created_at")
	assert.Empty(m.Courier, eo.Data.Attributes.Courier, "should not have courier")
	assert.Empty(m.CourierAccount, "should not have courier account")
	assert.Empty(m.Comments, "should not have comments")
	assert.Empty(m.Payment, "should not have payment")
	assert.Empty(m.PurchaseOrderNum, "should not have purchase order number")
	assert.Equal(m.Consumer, eo.Data.Attributes.Consumer, "should match the consumer")
	assert.Equal(m.Payer, eo.Data.Attributes.Payer, "should match the payer")
	assert.Equal(m.Purchaser, eo.Data.Attributes.Purchaser, "should match the purchaser")
	assert.ElementsMatch(m.Items, eo.Data.Attributes.Items, "should match the items")
	assert.NotEmpty(m.Key, "should not have empty key/id")
}

func testModelListSort(m []*model.OrderDoc, t *testing.T) {
	it, err := NewPairWiseIterator(m)
	if err != nil {
		t.Fatal(err)
	}
	assert := assert.New(t)
	for it.NextPair() {
		cm, nm := it.Pair()
		assert.Truef(
			nm.CreatedAt.Before(cm.CreatedAt),
			"date %s should be before %s",
			nm.CreatedAt.String(),
			cm.CreatedAt.String(),
		)
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

func convertFilterToQuery(s string) string {
	// parse filter logic
	// this needs to be done here since it is implemented in the service, not repository
	p, err := query.ParseFilterString(s)
	if err != nil {
		log.Printf("error parsing filter string %s", err)
		return s
	}
	str, err := query.GenAQLFilterStatement(&query.StatementParameters{Fmap: FMap, Filters: p, Doc: "s"})
	if err != nil {
		log.Printf("error generating AQL filter statement %s", err)
		return s
	}
	return str
}
