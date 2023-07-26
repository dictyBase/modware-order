package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dictyBase/apihelpers/aphgrpc"
	"github.com/dictyBase/arangomanager/query"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/message"
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
	ord.Data = &order.Order_Data{
		Type: s.GetResourceName(),
		Id:   mord.Key,
		Attributes: &order.OrderAttributes{
			CreatedAt:        aphgrpc.TimestampProto(mord.CreatedAt),
			UpdatedAt:        aphgrpc.TimestampProto(mord.UpdatedAt),
			Courier:          mord.Courier,
			CourierAccount:   mord.CourierAccount,
			Comments:         mord.Comments,
			Payment:          mord.Payment,
			PurchaseOrderNum: mord.PurchaseOrderNum,
			Status:           statusToEnum(mord.Status),
			Consumer:         mord.Consumer,
			Payer:            mord.Payer,
			Purchaser:        mord.Purchaser,
			Items:            mord.Items,
		},
	}

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
	ord.Data = &order.Order_Data{
		Type: s.GetResourceName(),
		Id:   adr.Key,
		Attributes: &order.OrderAttributes{
			CreatedAt:        aphgrpc.TimestampProto(adr.CreatedAt),
			UpdatedAt:        aphgrpc.TimestampProto(adr.UpdatedAt),
			Courier:          adr.Courier,
			CourierAccount:   adr.CourierAccount,
			Comments:         adr.Comments,
			Payment:          adr.Payment,
			PurchaseOrderNum: adr.PurchaseOrderNum,
			Status:           statusToEnum(adr.Status),
			Consumer:         adr.Consumer,
			Payer:            adr.Payer,
			Purchaser:        adr.Purchaser,
			Items:            adr.Items,
		},
	}
	s.publisher.Publish(s.Topics["orderCreate"], ord)

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
	ord.Data = &order.Order_Data{
		Type: s.GetResourceName(),
		Id:   mord.Key,
		Attributes: &order.OrderAttributes{
			CreatedAt:        aphgrpc.TimestampProto(mord.CreatedAt),
			UpdatedAt:        aphgrpc.TimestampProto(mord.UpdatedAt),
			Courier:          mord.Courier,
			CourierAccount:   mord.CourierAccount,
			Comments:         mord.Comments,
			Payment:          mord.Payment,
			PurchaseOrderNum: mord.PurchaseOrderNum,
			Status:           statusToEnum(mord.Status),
			Consumer:         mord.Consumer,
			Payer:            mord.Payer,
			Purchaser:        mord.Purchaser,
			Items:            mord.Items,
		},
	}
	s.publisher.Publish(s.Topics["orderUpdate"], ord) //nolint

	return ord, nil
}

func (s *OrderService) orderQueryWithFilter(
	ctx context.Context,
	params *order.ListParameters,
) (*order.OrderCollection, error) {
	oc := &order.OrderCollection{}
	p, err := query.ParseFilterString(params.Filter)
	if err != nil {
		return oc, fmt.Errorf("error parsing filter string: %s", err)
	}
	str, err := query.GenAQLFilterStatement(&query.StatementParameters{
		Fmap:    arangodb.FMap,
		Filters: p,
		Doc:     "s",
	})
	if err != nil {
		return oc, fmt.Errorf("error generating AQL filter statement: %s", err)
	}
	// if the parsed statement is empty FILTER, just return empty string
	if str == "FILTER " {
		str = ""
	}
	mc, err := s.repo.ListOrders(&order.ListParameters{
		Cursor: params.Cursor,
		Limit:  params.Limit,
		Filter: str,
	})
	if err != nil {
		return oc, aphgrpc.HandleGetError(ctx, err)
	}
	if len(mc) == 0 {
		return oc, aphgrpc.HandleNotFoundError(ctx, err)
	}
	var ocdata []*order.OrderCollection_Data
	for _, m := range mc {
		ocdata = append(ocdata, &order.OrderCollection_Data{
			Type: s.GetResourceName(),
			Id:   m.Key,
			Attributes: &order.OrderAttributes{
				CreatedAt:        aphgrpc.TimestampProto(m.CreatedAt),
				UpdatedAt:        aphgrpc.TimestampProto(m.UpdatedAt),
				Courier:          m.Courier,
				CourierAccount:   m.CourierAccount,
				Comments:         m.Comments,
				Payment:          m.Payment,
				PurchaseOrderNum: m.PurchaseOrderNum,
				Status:           statusToEnum(m.Status),
				Consumer:         m.Consumer,
				Payer:            m.Payer,
				Purchaser:        m.Purchaser,
				Items:            m.Items,
			},
		})
	}
	if len(ocdata) < int(params.Limit)-2 { // fewer results than limit
		oc.Data = ocdata
		oc.Meta = &order.Meta{Limit: params.Limit, Total: int64(len(ocdata))}

		return oc, nil
	}
	oc.Data = ocdata[:len(ocdata)-1]
	oc.Meta = &order.Meta{
		Limit:      params.Limit,
		NextCursor: genNextCursorVal(mc[len(mc)-1].CreatedAt),
		Total:      int64(len(ocdata)),
	}

	return oc, nil
}
		if err != nil {
			return oc, fmt.Errorf(
				"error generating AQL filter statement: %s",
				err,
			)
		}
		// if the parsed statement is empty FILTER, just return empty string
		if str == "FILTER " {
			str = ""
		}
		mc, err := s.repo.ListOrders(
			&order.ListParameters{Cursor: c, Limit: l, Filter: str},
		)
		if err != nil {
			return oc, aphgrpc.HandleGetError(ctx, err)
		}
		if len(mc) == 0 {
			return oc, aphgrpc.HandleNotFoundError(ctx, err)
		}
		var ocdata []*order.OrderCollection_Data
		for _, m := range mc {
			ocdata = append(ocdata, &order.OrderCollection_Data{
				Type: s.GetResourceName(),
				Id:   m.Key,
				Attributes: &order.OrderAttributes{
					CreatedAt:        aphgrpc.TimestampProto(m.CreatedAt),
					UpdatedAt:        aphgrpc.TimestampProto(m.UpdatedAt),
					Courier:          m.Courier,
					CourierAccount:   m.CourierAccount,
					Comments:         m.Comments,
					Payment:          m.Payment,
					PurchaseOrderNum: m.PurchaseOrderNum,
					Status:           statusToEnum(m.Status),
					Consumer:         m.Consumer,
					Payer:            m.Payer,
					Purchaser:        m.Purchaser,
					Items:            m.Items,
				},
			})
		}
		if len(ocdata) < int(l)-2 { // fewer results than limit
			oc.Data = ocdata
			oc.Meta = &order.Meta{Limit: l, Total: int64(len(ocdata))}
			return oc, nil
		}
		oc.Data = ocdata[:len(ocdata)-1]
		oc.Meta = &order.Meta{
			Limit:      l,
			NextCursor: genNextCursorVal(mc[len(mc)-1].CreatedAt),
			Total:      int64(len(ocdata)),
		}
	} else {
		mc, err := s.repo.ListOrders(&order.ListParameters{Cursor: c, Limit: l})
		if err != nil {
			return oc, aphgrpc.HandleGetError(ctx, err)
		}
		if len(mc) == 0 {
			return oc, aphgrpc.HandleNotFoundError(ctx, err)
		}
		var ocdata []*order.OrderCollection_Data
		for _, m := range mc {
			ocdata = append(ocdata, &order.OrderCollection_Data{
				Type: s.GetResourceName(),
				Id:   m.Key,
				Attributes: &order.OrderAttributes{
					CreatedAt:        aphgrpc.TimestampProto(m.CreatedAt),
					UpdatedAt:        aphgrpc.TimestampProto(m.UpdatedAt),
					Courier:          m.Courier,
					CourierAccount:   m.CourierAccount,
					Comments:         m.Comments,
					Payment:          m.Payment,
					PurchaseOrderNum: m.PurchaseOrderNum,
					Status:           statusToEnum(m.Status),
					Consumer:         m.Consumer,
					Payer:            m.Payer,
					Purchaser:        m.Purchaser,
					Items:            m.Items,
				},
			})
		}
		if len(ocdata) < int(l)-2 { // fewer results than limit
			oc.Data = ocdata
			oc.Meta = &order.Meta{Limit: l, Total: int64(len(ocdata))}
			return oc, nil
		}
		oc.Data = ocdata[:len(ocdata)-1]
		oc.Meta = &order.Meta{
			Limit:      l,
			NextCursor: genNextCursorVal(mc[len(mc)-1].CreatedAt),
			Total:      int64(len(ocdata)),
		}
	}
	return oc, nil
}

// LoadOrder handles the loading of an existing order
func (s *OrderService) LoadOrder(
	ctx context.Context,
	r *order.ExistingOrder,
) (*order.Order, error) {
	ord := &order.Order{}
	if err := r.Validate(); err != nil {
		return ord, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	m, err := s.repo.LoadOrder(r)
	if err != nil {
		return ord, aphgrpc.HandleInsertError(ctx, err)
	}
	if m.NotFound {
		return ord, aphgrpc.HandleNotFoundError(ctx, err)
	}
	ord.Data = &order.Order_Data{
		Type: s.GetResourceName(),
		Id:   m.Key,
		Attributes: &order.OrderAttributes{
			CreatedAt:        aphgrpc.TimestampProto(m.CreatedAt),
			UpdatedAt:        aphgrpc.TimestampProto(m.UpdatedAt),
			Courier:          m.Courier,
			CourierAccount:   m.CourierAccount,
			Comments:         m.Comments,
			Payment:          m.Payment,
			PurchaseOrderNum: m.PurchaseOrderNum,
			Status:           statusToEnum(m.Status),
			Consumer:         m.Consumer,
			Payer:            m.Payer,
			Purchaser:        m.Purchaser,
			Items:            m.Items,
		},
	}
	s.publisher.Publish(s.Topics["orderCreate"], ord)
	return ord, nil
}

// PrepareForOrder clears the database to prepare for loading data
func (s *OrderService) PrepareForOrder(
	ctx context.Context,
	r *empty.Empty,
) (*empty.Empty, error) {
	e := &empty.Empty{}
	if err := s.repo.ClearOrders(); err != nil {
		return e, aphgrpc.HandleGenericError(ctx, err)
	}
	return e, nil
}

// genNextCursorVal converts to epoch(https://en.wikipedia.org/wiki/Unix_time)
// in milliseconds
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
