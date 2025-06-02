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
	"github.com/stretchr/testify/require"
)

const (
	charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	seedRand   *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	gta        *testarango.TestArango
	collection = "stock_orders"
)

func stringWithCharset(length int, charset string) string {
	var byt []byte
	for i := 0; i < length; i++ {
		byt = append(
			byt,
			charset[seedRand.Intn(len(charset))],
		)
	}

	return string(byt)
}

func RandString(length int) string {
	return stringWithCharset(length, charSet)
}

func convertFilterToQuery(fstr string) string {
	// parse filter logic
	// this needs to be done here since it is implemented in the service, not repository
	pft, err := query.ParseFilterString(fstr)
	if err != nil {
		log.Printf("error parsing filter string %s", err)

		return fstr
	}
	str, err := query.GenAQLFilterStatement(
		&query.StatementParameters{Fmap: FMap, Filters: pft, Doc: "s"},
	)
	if err != nil {
		log.Printf("error generating AQL filter statement %s", err)

		return str
	}

	return str
}

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
				Status:           order.OrderStatus_IN_PREPARATION,
				Consumer:         consumer,
				Payer:            "dr.van.nostrand@gmail.com",
				Purchaser:        "dr.van.nostrand@gmail.com",
				Items:            []string{"DBS2109858", "DBP8349822"},
			},
		},
	}
}

/* func testModelListSort(assert *assert.Assertions, m []*model.OrderDoc) {
	itr, err := NewPairWiseIterator(m)
	assert.NoErrorf(err, "expect no error, received %s", err)
	for itr.NextPair() {
		cm, nm := itr.Pair()
		assert.Truef(
			nm.CreatedAt.Before(cm.CreatedAt),
			"date %s should be before %s",
			nm.CreatedAt.String(),
			cm.CreatedAt.String(),
		)
	}
} */

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
		dbh.Drop() //nolint
		log.Fatalf("unable to create collection %s %s", collection, err)
	}
	code := m.Run()
	dbh.Drop() //nolint
	os.Exit(code)
}

func TestAddOrder(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	// defer repo.ClearOrders() //nolint
	ntr := newTestOrder("art@vandelayindustries.com")
	mro, err := repo.AddOrder(ntr)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Equal(
		mro.Courier,
		ntr.Data.Attributes.Courier,
		"should match the courier",
	)
	require.Equal(
		mro.CourierAccount,
		ntr.Data.Attributes.CourierAccount,
		"should match the courier account",
	)
	require.Equal(
		mro.Comments,
		ntr.Data.Attributes.Comments,
		"should match the comments",
	)
	require.Equal(
		mro.Payment,
		ntr.Data.Attributes.Payment,
		"should match the payment",
	)
	require.Equal(
		mro.PurchaseOrderNum,
		ntr.Data.Attributes.PurchaseOrderNum,
		"should match the purchase order number",
	)
	require.Equal(
		mro.Status,
		ntr.Data.Attributes.Status.String(),
		"should match the status",
	)
	require.Equal(
		mro.Consumer,
		ntr.Data.Attributes.Consumer,
		"should match the consumer",
	)
	require.Equal(
		mro.Payer,
		ntr.Data.Attributes.Payer,
		"should match the payer",
	)
	require.Equal(
		mro.Purchaser,
		ntr.Data.Attributes.Purchaser,
		"should match the purchaser",
	)
	require.Equal(
		mro.Items,
		ntr.Data.Attributes.Items,
		"should match the items",
	)
	require.NotEmpty(mro.Key, "should not have empty key/id")
}

func TestGetOrder(t *testing.T) {
	t.Parallel()
	connP := getConnectParams()
	require := require.New(t)
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	// defer repo.ClearOrders() //nolint
	nrd := newTestOrder("art@vandelayindustries.com")
	mrd, err := repo.AddOrder(nrd)
	require.NoErrorf(err, "expect no error, received %s", err)
	grd, err := repo.GetOrder(mrd.Key)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Falsef(grd.NotFound, "expect the order %s to be found", mrd.Key)
	require.Equal(
		grd.Courier,
		nrd.Data.Attributes.Courier,
		"should match the courier",
	)
	require.Equal(
		grd.CourierAccount,
		nrd.Data.Attributes.CourierAccount,
		"should match the courier account",
	)
	require.Equal(
		grd.Comments,
		nrd.Data.Attributes.Comments,
		"should match the comments",
	)
	require.Equal(
		grd.Payment,
		nrd.Data.Attributes.Payment,
		"should match the payment",
	)
	require.Equal(
		grd.PurchaseOrderNum,
		nrd.Data.Attributes.PurchaseOrderNum,
		"should match the purchase order number",
	)
	require.Equal(
		grd.Status,
		nrd.Data.Attributes.Status.String(),
		"should match the status",
	)
	require.Equal(
		grd.Consumer,
		nrd.Data.Attributes.Consumer,
		"should match the consumer",
	)
	require.Equal(
		grd.Payer,
		nrd.Data.Attributes.Payer,
		"should match the payer",
	)
	require.Equal(
		grd.Purchaser,
		nrd.Data.Attributes.Purchaser,
		"should match the purchaser",
	)
	require.Equal(
		grd.Items,
		nrd.Data.Attributes.Items,
		"should match the items",
	)
	require.Len(grd.Items, 2, "should match length of two items")
	require.NotEmpty(grd.Key, "should not have empty key/id")
	require.True(
		mrd.CreatedAt.Equal(grd.CreatedAt),
		"should match created time of order",
	)
	require.True(
		mrd.UpdatedAt.Equal(grd.UpdatedAt),
		"should match updated time of order",
	)
	nre, err := repo.GetOrder("1")
	require.NoErrorf(err, "expect no error, received %s", err)
	require.True(nre.NotFound, "entry should not exist")
}

func TestEditOrder(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	// defer repo.ClearOrders() //nolint
	no := newTestOrder("art@vandelayindustries.com")
	mrd, err := repo.AddOrder(no)
	require.NoErrorf(err, "expect no error, received %s", err)
	testData := &order.OrderUpdate{Data: &order.OrderUpdate_Data{
		Type: "order",
		Id:   mrd.Key,
		Attributes: &order.OrderUpdateAttributes{
			Courier:          "UPS",
			CourierAccount:   "99999999",
			Comments:         "This is an updated test comment",
			Payment:          "Check",
			PurchaseOrderNum: "33333333",
			Items:            []string{"xyz", "abc"},
			Status:           order.OrderStatus_GROWING,
		},
	}}
	edr, err := repo.EditOrder(testData)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Equal(
		edr.Courier,
		testData.Data.Attributes.Courier,
		"should match the new courier",
	)
	require.Equal(
		edr.CourierAccount,
		testData.Data.Attributes.CourierAccount,
		"should match the new courier account",
	)
	require.Equal(
		edr.Comments,
		testData.Data.Attributes.Comments,
		"should match the new comments",
	)
	require.Equal(
		edr.Payment,
		testData.Data.Attributes.Payment,
		"should match the new payment",
	)
	require.Equal(
		edr.PurchaseOrderNum,
		testData.Data.Attributes.PurchaseOrderNum,
		"should match the new purchase order number",
	)
	require.ElementsMatch(
		edr.Items,
		testData.Data.Attributes.Items,
		"should match the new items",
	)
	require.Equal(
		edr.Status,
		testData.Data.Attributes.Status.String(),
		"should match the new status",
	)
	grd, err := repo.GetOrder(mrd.Key)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Equal(
		grd.Payer,
		mrd.Payer,
		"should match the already existing payer",
	)
	require.Equal(edr.Courier, grd.Courier, "should match the new courier")
	oed := &order.OrderUpdate{
		Data: &order.OrderUpdate_Data{
			Type: "order",
			Id:   "1",
			Attributes: &order.OrderUpdateAttributes{
				Comments: "This is an updated test comment",
			},
		},
	}
	ee, err := repo.EditOrder(oed)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.True(ee.NotFound, "entry should not exist")
}

func TestListOrders(t *testing.T) { //nolint:paralleltest
	require := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.NoError(
		repo.ClearOrders(),
		"failed to clear orders at start of TestListOrders",
	)
	// defer repo.ClearOrders() //nolint
	for i := 1; i <= 15; i++ {
		no := newTestOrder(
			fmt.Sprintf("%s@kramericaindustries.com", RandString(10)),
		)
		_, err = repo.AddOrder(no)
		require.NoErrorf(err, "expect no error, received %s", err)
	}
	lrd, err := repo.ListOrders(&order.ListParameters{Limit: 4})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Len(lrd, 5, "should match the provided limit number + 1")
	for _, order := range lrd {
		require.Equal("FedEx", order.Courier, "should match the courier")
		require.NotEmpty(order.Key, "should not have empty key/id")
	}
	require.NotEqual(
		lrd[0].Consumer,
		lrd[1].Consumer,
		"should have different consumers",
	)
	ti := toTimestamp(lrd[len(lrd)-1].CreatedAt)
	lo2, err := repo.ListOrders(&order.ListParameters{Cursor: ti, Limit: 4})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Len(lo2, 5, "should match the provided limit number + 1")
	require.NotEqual(
		lo2[0].Consumer,
		lo2[1].Consumer,
		"should have different consumers",
	)
	ti2 := toTimestamp(lo2[len(lo2)-1].CreatedAt)
	lo3, err := repo.ListOrders(&order.ListParameters{Cursor: ti2, Limit: 4})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Len(lo3, 5, "should match the provided limit number + 1")
	ti3 := toTimestamp(lo3[len(lo3)-1].CreatedAt)
	lo4, err := repo.ListOrders(&order.ListParameters{Cursor: ti3, Limit: 4})
	require.NoErrorf(err, "expect no error, received %s", err)
	// Depends on how the timestamps of individual element is created,
	// the query can fetch three or more elements. Since it's going
	// towards the end of the list, the returned element number varies
	// depending on which particular index in the list it matches
	require.GreaterOrEqual(
		len(lo4),
		3,
		"should at least bring last three results",
	)
	sfd, err := repo.ListOrders(&order.ListParameters{
		Limit:  100,
		Filter: convertFilterToQuery("courier===FedEx"),
	})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Len(sfd, 15, "should list all 15 orders")
	scd, err := repo.ListOrders(&order.ListParameters{
		Cursor: toTimestamp(sfd[5].CreatedAt),
		Limit:  100,
		Filter: convertFilterToQuery("courier===FedEx"),
	})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.GreaterOrEqual(len(scd), 10, "should list at least last 10 orders")
	snd, err := repo.ListOrders(&order.ListParameters{
		Cursor: toTimestamp(sfd[5].CreatedAt),
		Limit:  100,
		Filter: convertFilterToQuery("courier===UPS"),
	})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Empty(snd, "should list last no UPS orders")
}

func TestLoadOrder(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	// defer repo.ClearOrders() //nolint
	tme, _ := time.Parse("2006-01-02 15:04:05", "2010-03-30 14:40:58")
	eod := &order.ExistingOrder{
		Data: &order.ExistingOrder_Data{
			Type: "order",
			Attributes: &order.ExistingOrderAttributes{
				CreatedAt: aphgrpc.TimestampProto(tme),
				UpdatedAt: aphgrpc.TimestampProto(tme),
				Purchaser: "super@c.org",
				Items:     []string{"DBS2109858", "DBP8349822"},
			},
		},
	}
	mrd, err := repo.LoadOrder(eod)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.True(mrd.CreatedAt.Equal(tme), "should match created_at")
	require.True(mrd.UpdatedAt.Equal(tme), "should match updated_at")
	require.Equal(
		mrd.Purchaser,
		eod.Data.Attributes.Purchaser,
		"should match the purchaser",
	)
	require.ElementsMatch(
		mrd.Items,
		eod.Data.Attributes.Items,
		"should match the items",
	)
	require.NotEmpty(mrd.Key, "should not have empty key/id")
}

//nolint:paralleltest
func TestClearOrders(t *testing.T) {
	require := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	require.NoErrorf(err, "expect no error, received %s", err)
	require.NoError(
		repo.ClearOrders(),
		"failed to clear orders at start of TestClearOrders",
	)
	// add 15 new test orders
	for i := 1; i <= 15; i++ {
		no := newTestOrder(
			fmt.Sprintf("%s@kramericaindustries.com", RandString(10)),
		)
		_, err = repo.AddOrder(no)
		require.NoErrorf(err, "expect no error, received %s", err)
	}
	lo, err := repo.ListOrders(&order.ListParameters{Limit: 100})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Len(
		lo,
		15,
		"should have exactly 15 orders after adding them to a cleared collection",
	)
	err = repo.ClearOrders()
	require.NoErrorf(err, "expect no error, received %s", err)
	lo2, err := repo.ListOrders(&order.ListParameters{Limit: 100})
	require.NoErrorf(err, "expect no error, received %s", err)
	require.Empty(lo2, "should not list any orders after final clear")
}
