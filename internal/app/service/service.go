package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dictyBase/apihelpers/aphgrpc"
	"github.com/dictyBase/arangomanager/query"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/message"
	"github.com/dictyBase/modware-order/internal/model"
	"github.com/dictyBase/modware-order/internal/repository"
	"github.com/dictyBase/modware-order/internal/repository/arangodb"
	"github.com/golang/protobuf/ptypes/empty"
)

const Divider = 1000000

// OrderService is the container for managing order service definition.
type OrderService struct {
	order.UnimplementedOrderServiceServer
	*aphgrpc.Service
	repo      repository.OrderRepository
	publisher message.Publisher
}

func defaultOptions() *aphgrpc.ServiceOptions {
	return &aphgrpc.ServiceOptions{Resource: "order"}
}

// NewOrderService is the constructor for creating a new instance of
// OrderService.
func NewOrderService(
	repo repository.OrderRepository,
	pub message.Publisher,
	opt ...aphgrpc.Option,
) *OrderService {
	so := defaultOptions()
	for _, optfn := range opt {
		optfn(so)
	}
	srv := &aphgrpc.Service{}
	aphgrpc.AssignFieldsToStructs(so, srv)

	return &OrderService{
		Service:   srv,
		repo:      repo,
		publisher: pub,
	}
}

// GetOrder handles getting an order by ID.
func (s *OrderService) GetOrder(
	ctx context.Context,
	rdr *order.OrderId,
) (*order.Order, error) {
	ord := &order.Order{}
	if err := rdr.Validate(); err != nil {
		return ord, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	mord, err := s.repo.GetOrder(rdr.Id)
	if err != nil {
		return ord, aphgrpc.HandleGetError(ctx, err)
	}
	if mord.NotFound {
		return ord, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ord.Data = s.orderData(mord)

	return ord, nil
}

// CreateOrder handles the creation of a new order.
func (s *OrderService) CreateOrder(
	ctx context.Context,
	rdr *order.NewOrder,
) (*order.Order, error) {
	ord := &order.Order{}
	if err := rdr.Validate(); err != nil {
		return ord, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	adr, err := s.repo.AddOrder(rdr)
	if err != nil {
		return ord, aphgrpc.HandleInsertError(ctx, err)
	}
	if adr.NotFound {
		return ord, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ord.Data = s.orderData(adr)
	s.publisher.Publish(s.Topics["orderCreate"], ord) //nolint:errcheck

	return ord, nil
}

// UpdateOrder handles updating an existing order.
func (s *OrderService) UpdateOrder(
	ctx context.Context,
	urd *order.OrderUpdate,
) (*order.Order, error) {
	ord := &order.Order{}
	if err := urd.Validate(); err != nil {
		return ord, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	mord, err := s.repo.EditOrder(urd)
	if err != nil {
		return ord, aphgrpc.HandleUpdateError(ctx, err)
	}
	if mord.NotFound {
		return ord, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ord.Data = s.orderData(mord)
	s.publisher.Publish(s.Topics["orderUpdate"], ord) //nolint

	return ord, nil
}

func (s *OrderService) orderQueryWithFilter(
	ctx context.Context,
	params *order.ListParameters,
) (*order.OrderCollection, error) {
	orc := &order.OrderCollection{}
	pfs, err := query.ParseFilterString(params.Filter)
	if err != nil {
		return orc, fmt.Errorf("error parsing filter string: %s", err)
	}
	str, err := query.GenAQLFilterStatement(&query.StatementParameters{
		Fmap:    arangodb.FMap,
		Filters: pfs,
		Doc:     "s",
	})
	if err != nil {
		return orc, fmt.Errorf("error generating AQL filter statement: %s", err)
	}
	// if the parsed statement is empty FILTER, just return empty string
	if str == "FILTER " {
		str = ""
	}
	mlo, err := s.repo.ListOrders(&order.ListParameters{
		Cursor: params.Cursor,
		Limit:  params.Limit,
		Filter: str,
	})
	if err != nil {
		return orc, aphgrpc.HandleGetError(ctx, err)
	}
	if len(mlo) == 0 {
		return orc, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ocdata := make([]*order.OrderCollection_Data, 0)
	for _, item := range mlo {
		ocdata = append(ocdata, &order.OrderCollection_Data{
			Type: s.GetResourceName(),
			Id:   item.Key,
			Attributes: &order.OrderAttributes{
				CreatedAt:        aphgrpc.TimestampProto(item.CreatedAt),
				UpdatedAt:        aphgrpc.TimestampProto(item.UpdatedAt),
				Courier:          item.Courier,
				CourierAccount:   item.CourierAccount,
				Comments:         item.Comments,
				Payment:          item.Payment,
				PurchaseOrderNum: item.PurchaseOrderNum,
				Status:           statusToEnum(item.Status),
				Consumer:         item.Consumer,
				Payer:            item.Payer,
				Purchaser:        item.Purchaser,
				Items:            item.Items,
			},
		})
	}
	if len(ocdata) < int(params.Limit)-2 { // fewer results than limit
		orc.Data = ocdata
		orc.Meta = &order.Meta{Limit: params.Limit, Total: int64(len(ocdata))}

		return orc, nil
	}
	orc.Data = ocdata[:len(ocdata)-1]
	orc.Meta = &order.Meta{
		Limit:      params.Limit,
		NextCursor: genNextCursorVal(mlo[len(mlo)-1].CreatedAt),
		Total:      int64(len(ocdata)),
	}

	return orc, nil
}

func (s *OrderService) orderQueryWithoutFilter(
	ctx context.Context,
	params *order.ListParameters,
) (*order.OrderCollection, error) {
	orc := &order.OrderCollection{}
	mlo, err := s.repo.ListOrders(&order.ListParameters{
		Cursor: params.Cursor,
		Limit:  params.Limit,
	})
	if err != nil {
		return orc, aphgrpc.HandleGetError(ctx, err)
	}
	if len(mlo) == 0 {
		return orc, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ocdata := make([]*order.OrderCollection_Data, 0)
	for _, item := range mlo {
		ocdata = append(ocdata, &order.OrderCollection_Data{
			Type: s.GetResourceName(),
			Id:   item.Key,
			Attributes: &order.OrderAttributes{
				CreatedAt:        aphgrpc.TimestampProto(item.CreatedAt),
				UpdatedAt:        aphgrpc.TimestampProto(item.UpdatedAt),
				Courier:          item.Courier,
				CourierAccount:   item.CourierAccount,
				Comments:         item.Comments,
				Payment:          item.Payment,
				PurchaseOrderNum: item.PurchaseOrderNum,
				Status:           statusToEnum(item.Status),
				Consumer:         item.Consumer,
				Payer:            item.Payer,
				Purchaser:        item.Purchaser,
				Items:            item.Items,
			},
		})
	}
	if len(ocdata) < int(params.Limit)-2 { // fewer results than limit
		orc.Data = ocdata
		orc.Meta = &order.Meta{Limit: params.Limit, Total: int64(len(ocdata))}

		return orc, nil
	}
	orc.Data = ocdata[:len(ocdata)-1]
	orc.Meta = &order.Meta{
		Limit:      params.Limit,
		NextCursor: genNextCursorVal(mlo[len(mlo)-1].CreatedAt),
		Total:      int64(len(ocdata)),
	}

	return orc, nil
}

// ListOrders lists all existing orders.
func (s *OrderService) ListOrders(
	ctx context.Context,
	params *order.ListParameters,
) (*order.OrderCollection, error) {
	orc := &order.OrderCollection{}
	lmt := params.Limit
	if params.Limit == 0 {
		lmt = 10
	}
	if len(params.Filter) > 0 {
		orc, err := s.orderQueryWithFilter(ctx, &order.ListParameters{
			Limit:  lmt,
			Cursor: params.Cursor,
			Filter: params.Filter,
		})
		if err != nil {
			return orc, err
		}
	} else {
		orc, err := s.orderQueryWithoutFilter(ctx, &order.ListParameters{
			Limit:  lmt,
			Cursor: params.Cursor,
		})
		if err != nil {
			return orc, err
		}
	}

	return orc, nil
}

// LoadOrder handles the loading of an existing order.
func (s *OrderService) LoadOrder(
	ctx context.Context,
	rxo *order.ExistingOrder,
) (*order.Order, error) {
	ord := &order.Order{}
	if err := rxo.Validate(); err != nil {
		return ord, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	mlrd, err := s.repo.LoadOrder(rxo)
	if err != nil {
		return ord, aphgrpc.HandleInsertError(ctx, err)
	}
	if mlrd.NotFound {
		return ord, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ord.Data = s.orderData(mlrd)
	s.publisher.Publish(s.Topics["orderCreate"], ord) //nolint:errcheck

	return ord, nil
}

// PrepareForOrder clears the database to prepare for loading data.
func (s *OrderService) PrepareForOrder(
	ctx context.Context,
	rmt *empty.Empty,
) (*empty.Empty, error) {
	e := &empty.Empty{}
	if err := s.repo.ClearOrders(); err != nil {
		return e, aphgrpc.HandleGenericError(ctx, err)
	}

	return e, nil
}

func (s *OrderService) orderData(ord *model.OrderDoc) *order.Order_Data {
	return &order.Order_Data{
		Type: s.GetResourceName(),
		Id:   ord.Key,
		Attributes: &order.OrderAttributes{
			CreatedAt:        aphgrpc.TimestampProto(ord.CreatedAt),
			UpdatedAt:        aphgrpc.TimestampProto(ord.UpdatedAt),
			Courier:          ord.Courier,
			CourierAccount:   ord.CourierAccount,
			Comments:         ord.Comments,
			Payment:          ord.Payment,
			PurchaseOrderNum: ord.PurchaseOrderNum,
			Status:           statusToEnum(ord.Status),
			Consumer:         ord.Consumer,
			Payer:            ord.Payer,
			Purchaser:        ord.Purchaser,
			Items:            ord.Items,
		},
	}
}

// genNextCursorVal converts to epoch(https://en.wikipedia.org/wiki/Unix_time)
// in milliseconds.
func genNextCursorVal(t time.Time) int64 {
	return t.UnixNano() / Divider
}

func statusToEnum(status string) order.OrderStatus {
	switch status {
	case "Shipped":
		return order.OrderStatus_SHIPPED
	case "Cancelled":
		return order.OrderStatus_CANCELLED
	case "Growing":
		return order.OrderStatus_GROWING
	default:
		break
	}

	return order.OrderStatus_IN_PREPARATION
}
