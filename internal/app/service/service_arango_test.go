package service

import (
	"context"
	"net"
	"testing"

	"github.com/dictyBase/arangomanager/testarango"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/repository/arangodb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	arangoTestCollection = "stock_order"
	fedexCourier         = "FedEx"
	upsCourier           = "UPS"
	unmatchedCourier     = "DHL"
)

// arangoTestEnv serves the order service from a disposable ArangoDB
// database over an in-process gRPC connection.
type arangoTestEnv struct {
	client order.OrderServiceClient
	pub    *stubPublisher
	tra    *testarango.TestArango
	assert *require.Assertions
}

// newArangoTestEnv creates a database, wires the real repository into the
// service and exposes it through a bufconn client.
func newArangoTestEnv(t *testing.T) *arangoTestEnv {
	t.Helper()
	assert := require.New(t)
	tra, err := testarango.NewTestArangoFromEnv(true)
	assert.NoErrorf(err, "expect no error creating a test database")
	repo, err := arangodb.NewOrderRepo(tra.ConnectParams, arangoTestCollection)
	assert.NoErrorf(err, "expect no error connecting to the repository")
	pub := &stubPublisher{}
	srv := grpc.NewServer()
	order.RegisterOrderServiceServer(srv, newTestService(repo, pub))
	lis := bufconn.Listen(1024 * 1024)
	go func() {
		// the returned error is gRPC's ErrServerStopped on cleanup; the
		// goroutine can outlive the test, so it must not log
		_ = srv.Serve(lis)
	}()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	assert.NoErrorf(err, "expect no error creating a grpc client")
	t.Cleanup(func() {
		if dbh, derr := tra.DB(tra.Database); derr == nil {
			_ = dbh.Drop()
		}
		_ = conn.Close()
		_ = lis.Close()
		srv.Stop()
	})

	return &arangoTestEnv{
		client: order.NewOrderServiceClient(conn),
		pub:    pub,
		tra:    tra,
		assert: assert,
	}
}

// seedOrders stores count orders carrying the given courier.
func (e *arangoTestEnv) seedOrders(t *testing.T, count int, courier string) {
	t.Helper()
	for idx := range count {
		input := newOrderInput()
		input.Data.Attributes.Courier = courier
		_, err := e.client.CreateOrder(context.Background(), input)
		e.assert.NoErrorf(
			err,
			"expect no error creating order %d, received %s",
			idx,
			err,
		)
	}
}

// listAllPages walks the collection page by page and returns how often
// each order was seen together with the number of pages consumed. It
// stops at maxPages so a pagination loop fails the test instead of
// hanging it.
func (e *arangoTestEnv) listAllPages(
	t *testing.T,
	limit int64,
	maxPages int,
) (map[string]int, int) {
	t.Helper()
	visited := map[string]int{}
	cursor := int64(0)
	for page := 1; page <= maxPages; page++ {
		got, err := e.client.ListOrders(
			context.Background(),
			&order.ListParameters{Limit: limit, Cursor: cursor},
		)
		e.assert.NoErrorf(err, "expect no error listing page %d, received %s", page, err)
		for _, data := range got.Data {
			visited[data.Id]++
		}
		if got.Meta.GetNextCursor() == 0 {
			return visited, page
		}
		cursor = got.Meta.GetNextCursor()
	}

	return visited, maxPages
}

// seedOneOrder stores a single order and returns its key.
func (e *arangoTestEnv) seedOneOrder(t *testing.T) string {
	t.Helper()
	created, err := e.client.CreateOrder(context.Background(), newOrderInput())
	e.assert.NoErrorf(err, "expect no error, received %s", err)

	return created.Data.Id
}

func TestCreateOrderStoresReturnedOrder(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	created, err := env.client.CreateOrder(context.Background(), newOrderInput())
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(t, orderResourceType, created.Data.Type)
	require.NotEmpty(t, created.Data.Id, "should have a generated key")
	require.Equal(t, fedexCourier, created.Data.Attributes.Courier)
	require.Equal(t, []string{createTopic}, env.pub.subjects)
	found, err := env.client.GetOrder(
		context.Background(),
		&order.OrderId{Id: created.Data.Id},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(t, created.Data.Id, found.Data.Id, "should be retrievable")
}

func TestCreateOrderRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	_, err := env.client.CreateOrder(
		context.Background(),
		&order.NewOrder{Data: &order.NewOrder_Data{Type: orderResourceType}},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetOrderReturnsStoredOrder(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 1, fedexCourier)
	listed, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, listed.Data, 1)
	got, err := env.client.GetOrder(
		context.Background(),
		&order.OrderId{Id: listed.Data[0].Id},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(t, fedexCourier, got.Data.Attributes.Courier)
	require.Equal(t, order.OrderStatus_IN_PREPARATION, got.Data.Attributes.Status)
}

func TestGetOrderUnknownKeyReturnsNotFound(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	_, err := env.client.GetOrder(
		context.Background(),
		&order.OrderId{Id: "not-a-stored-order"},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateOrderUpdatesStoredOrder(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 1, fedexCourier)
	listed, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	key := listed.Data[0].Id
	updated, err := env.client.UpdateOrder(
		context.Background(),
		&order.OrderUpdate{
			Data: &order.OrderUpdate_Data{
				Type: orderResourceType,
				Id:   key,
				Attributes: &order.OrderUpdateAttributes{
					Courier: upsCourier,
				},
			},
		},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(t, upsCourier, updated.Data.Attributes.Courier)
	require.Contains(t, env.pub.subjects, updateTopic)
}

func TestUpdateOrderUnknownKeyReturnsNotFound(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	_, err := env.client.UpdateOrder(
		context.Background(),
		&order.OrderUpdate{
			Data: &order.OrderUpdate_Data{
				Type: orderResourceType,
				Id:   "not-a-stored-order",
				Attributes: &order.OrderUpdateAttributes{
					Courier: upsCourier,
				},
			},
		},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestListOrdersReturnsAllStoredOrders(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 3, fedexCourier)
	got, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 3)
	require.Equal(t, int64(3), got.Meta.GetTotal())
	require.Zero(t, got.Meta.GetNextCursor(), "last page carries no cursor")
}

func TestListOrdersWithFilterReturnsMatchingOrders(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 3, fedexCourier)
	env.seedOrders(t, 2, upsCourier)
	got, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10, Filter: "courier===" + fedexCourier},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 3, "should only return the matching orders")
	for _, data := range got.Data {
		require.Equal(t, fedexCourier, data.Attributes.Courier)
	}
}

func TestListOrdersWithUnmatchedFilterReturnsNotFound(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 2, fedexCourier)
	_, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10, Filter: "courier===" + unmatchedCourier},
	)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestListOrdersPaginatesWithoutGapsOrDuplicates(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 15, fedexCourier)
	visited, pages := env.listAllPages(t, 5, 10)
	require.Len(t, visited, 15, "every stored order should be paged through")
	for key, count := range visited {
		require.Equal(t, 1, count, "order %s should be returned once", key)
	}
	require.Equal(t, 3, pages, "15 orders at 5 per page is three pages")
}

func TestListOrdersLastPageEqualToLimitReturnsFullPage(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 15, fedexCourier)
	got, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 15},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 15, "the whole collection fits in one page")
	require.Zero(t, got.Meta.GetNextCursor(), "there is no next page")
}

func TestListOrdersLastPageTwoBelowLimitReturnsFullPage(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 13, fedexCourier)
	got, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 15},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 13, "the whole collection fits in one page")
	require.Zero(t, got.Meta.GetNextCursor(), "there is no next page")
}

func TestListOrdersSingleItemPagesTerminate(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 5, fedexCourier)
	visited, pages := env.listAllPages(t, 1, 12)
	require.Len(t, visited, 5, "every stored order should be paged through")
	for key, count := range visited {
		require.Equal(t, 1, count, "order %s should be returned once", key)
	}
	require.Equal(t, 5, pages, "5 orders at 1 per page is five pages")
}

func TestListOrdersTotalMatchesReturnedRows(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 15, fedexCourier)
	got, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 5},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(
		t,
		int64(len(got.Data)),
		got.Meta.GetTotal(),
		"total should describe the rows the caller received",
	)
}

func TestLoadOrderStoresOrder(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	loaded, err := env.client.LoadOrder(context.Background(), existingOrderInput())
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.NotEmpty(t, loaded.Data.Id, "should have a generated key")
	require.Equal(t, []string{createTopic}, env.pub.subjects)
	listed, err := env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, listed.Data, 1, "the loaded order should be stored")
	require.Equal(t, loaded.Data.Id, listed.Data[0].Id)
}

func TestLoadOrderRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	_, err := env.client.LoadOrder(
		context.Background(),
		&order.ExistingOrder{
			Data: &order.ExistingOrder_Data{Type: orderResourceType},
		},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestPrepareForOrderClearsStoredOrders(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	env.seedOrders(t, 2, fedexCourier)
	_, err := env.client.PrepareForOrder(context.Background(), &emptypb.Empty{})
	require.NoErrorf(t, err, "expect no error, received %s", err)
	_, err = env.client.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"collection should be empty",
	)
}

func TestUpdateOrderWithoutAttributesReturnsInvalidArgument(t *testing.T) {
	t.Parallel()
	env := newArangoTestEnv(t)
	seeded := env.seedOneOrder(t)
	_, err := env.client.UpdateOrder(
		context.Background(),
		&order.OrderUpdate{
			Data: &order.OrderUpdate_Data{
				Type: orderResourceType,
				Id:   seeded,
			},
		},
	)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
