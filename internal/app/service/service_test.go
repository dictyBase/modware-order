package service

import (
	"context"
	"errors"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/aphgrpc"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/message"
	"github.com/dictyBase/modware-order/internal/model"
	"github.com/dictyBase/modware-order/internal/repository"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	orderResourceType = "order"
	testOrderKey      = "order-key"
	testPurchaser     = "someone@example.org"
	createTopic       = "OrderService.Create"
	updateTopic       = "OrderService.Update"
)

// stubRepository serves canned order documents so the service can be
// exercised without a database.
type stubRepository struct {
	doc  *model.OrderDoc
	docs []*model.OrderDoc
	err  error
}

func (s *stubRepository) GetOrder(_ string) (*model.OrderDoc, error) {
	return s.doc, s.err
}

func (s *stubRepository) AddOrder(_ *order.NewOrder) (*model.OrderDoc, error) {
	return s.doc, s.err
}

func (s *stubRepository) EditOrder(_ *order.OrderUpdate) (*model.OrderDoc, error) {
	return s.doc, s.err
}

func (s *stubRepository) ListOrders(
	_ *order.ListParameters,
) ([]*model.OrderDoc, error) {
	return s.docs, s.err
}

func (s *stubRepository) LoadOrder(
	_ *order.ExistingOrder,
) (*model.OrderDoc, error) {
	return s.doc, s.err
}

func (s *stubRepository) ClearOrders() error {
	return s.err
}

func (s *stubRepository) Autocomplete(
	_ string,
	_ int,
) ([]*repository.Suggestion, error) {
	return nil, s.err
}

// stubPublisher records the subjects it was asked to publish.
type stubPublisher struct {
	subjects []string
	err      error
}

func (p *stubPublisher) Publish(subject string, _ *order.Order) error {
	p.subjects = append(p.subjects, subject)

	return p.err
}

func (p *stubPublisher) Close() error {
	return nil
}

func newTestService(
	repo repository.OrderRepository,
	pub message.Publisher,
) *OrderService {
	return NewOrderService(repo, pub, aphgrpc.TopicsOption(map[string]string{
		"orderCreate": createTopic,
		"orderUpdate": updateTopic,
	}))
}

func missingDoc() *model.OrderDoc {
	return &model.OrderDoc{NotFound: true}
}

func orderDocs(count int) []*model.OrderDoc {
	docs := make([]*model.OrderDoc, 0, count)
	for i := range count {
		docs = append(docs, &model.OrderDoc{
			DocumentMeta: driver.DocumentMeta{Key: testOrderKey},
			CreatedAt:    time.Now().Add(time.Duration(i) * time.Second),
			UpdatedAt:    time.Now(),
			Courier:      "FedEx",
			Status:       "Shipped",
			Purchaser:    testPurchaser,
		})
	}

	return docs
}

func newOrderInput() *order.NewOrder {
	return &order.NewOrder{
		Data: &order.NewOrder_Data{
			Type: orderResourceType,
			Attributes: &order.NewOrderAttributes{
				Courier:          "FedEx",
				CourierAccount:   "9912378999",
				Payment:          "Credit card",
				PurchaseOrderNum: "38975932199",
				Status:           order.OrderStatus_IN_PREPARATION,
				Consumer:         testPurchaser,
				Payer:            testPurchaser,
				Purchaser:        testPurchaser,
				Items:            []string{"DBS2109858"},
			},
		},
	}
}

func orderUpdateInput() *order.OrderUpdate {
	return &order.OrderUpdate{
		Data: &order.OrderUpdate_Data{
			Type: orderResourceType,
			Id:   testOrderKey,
			Attributes: &order.OrderUpdateAttributes{
				Courier: "UPS",
			},
		},
	}
}

func existingOrderInput() *order.ExistingOrder {
	return &order.ExistingOrder{
		Data: &order.ExistingOrder_Data{
			Type: orderResourceType,
			Attributes: &order.ExistingOrderAttributes{
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
				Purchaser: testPurchaser,
				Items:     []string{"DBS2109858"},
			},
		},
	}
}

func TestGetOrderMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{doc: missingDoc()}, &stubPublisher{})
	_, err := svc.GetOrder(
		context.Background(),
		&order.OrderId{Id: "does-not-exist"},
	)
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"missing order should be reported as NotFound, got %v",
		err,
	)
}

func TestUpdateOrderMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{doc: missingDoc()}, &stubPublisher{})
	_, err := svc.UpdateOrder(context.Background(), orderUpdateInput())
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"missing order should be reported as NotFound, got %v",
		err,
	)
}

func TestLoadOrderMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{doc: missingDoc()}, &stubPublisher{})
	_, err := svc.LoadOrder(context.Background(), existingOrderInput())
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"missing order should be reported as NotFound, got %v",
		err,
	)
}

func TestListOrdersEmptyWithoutFilterReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	_, err := svc.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"empty result should be reported as NotFound, got %v",
		err,
	)
}

func TestListOrdersEmptyWithFilterReturnsNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	_, err := svc.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10, Filter: "courier===FedEx"},
	)
	require.Equal(
		t,
		codes.NotFound,
		status.Code(err),
		"empty filtered result should be reported as NotFound, got %v",
		err,
	)
}

func TestListOrdersWithoutFilterReturnsRepositoryDocuments(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{docs: orderDocs(3)}, &stubPublisher{})
	got, err := svc.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 3, "should return every stored order")
	require.Equal(t, int64(3), got.Meta.GetTotal(), "should report the total")
}

func TestListOrdersWithFilterReturnsRepositoryDocuments(t *testing.T) {
	t.Parallel()
	svc := newTestService(&stubRepository{docs: orderDocs(3)}, &stubPublisher{})
	got, err := svc.ListOrders(
		context.Background(),
		&order.ListParameters{Limit: 10, Filter: "courier===FedEx"},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Len(t, got.Data, 3, "should return every matching order")
	require.Equal(t, int64(3), got.Meta.GetTotal(), "should report the total")
}

func TestGetOrderReturnsMappedAttributes(t *testing.T) {
	t.Parallel()
	doc := orderDocs(1)[0]
	svc := newTestService(&stubRepository{doc: doc}, &stubPublisher{})
	got, err := svc.GetOrder(
		context.Background(),
		&order.OrderId{Id: testOrderKey},
	)
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(t, orderResourceType, got.Data.Type, "should carry the resource name")
	require.Equal(t, testOrderKey, got.Data.Id, "should carry the document key")
	require.Equal(t, "FedEx", got.Data.Attributes.Courier)
	require.Equal(
		t,
		order.OrderStatus_SHIPPED,
		got.Data.Attributes.Status,
		"should map the stored status to its enum",
	)
}

func TestCreateOrderPublishFailureReturnsInternalError(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{err: errors.New("messaging server unreachable")}
	svc := newTestService(&stubRepository{doc: orderDocs(1)[0]}, pub)
	got, err := svc.CreateOrder(context.Background(), newOrderInput())
	require.Equal(
		t,
		codes.Internal,
		status.Code(err),
		"publish failure should fail the request, got %v",
		err,
	)
	require.Equal(t, []string{createTopic}, pub.subjects)
	require.NotNil(t, got, "should still return the created order")
}

func TestUpdateOrderPublishFailureReturnsInternalError(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{err: errors.New("messaging server unreachable")}
	svc := newTestService(&stubRepository{doc: orderDocs(1)[0]}, pub)
	got, err := svc.UpdateOrder(context.Background(), orderUpdateInput())
	require.Equal(
		t,
		codes.Internal,
		status.Code(err),
		"publish failure should fail the request, got %v",
		err,
	)
	require.Equal(t, []string{updateTopic}, pub.subjects)
	require.NotNil(t, got, "should still return the updated order")
}

func TestLoadOrderPublishFailureReturnsInternalError(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{err: errors.New("messaging server unreachable")}
	svc := newTestService(&stubRepository{doc: orderDocs(1)[0]}, pub)
	got, err := svc.LoadOrder(context.Background(), existingOrderInput())
	require.Equal(
		t,
		codes.Internal,
		status.Code(err),
		"publish failure should fail the request, got %v",
		err,
	)
	require.Equal(t, []string{createTopic}, pub.subjects)
	require.NotNil(t, got, "should still return the loaded order")
}

func TestCreateOrderPublishesToCreateTopic(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{}
	svc := newTestService(&stubRepository{doc: orderDocs(1)[0]}, pub)
	_, err := svc.CreateOrder(context.Background(), newOrderInput())
	require.NoErrorf(t, err, "expect no error, received %s", err)
	require.Equal(
		t,
		[]string{createTopic},
		pub.subjects,
		"should publish on the create topic",
	)
}
