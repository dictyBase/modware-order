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

// OrderService is the container for managing order service definition
type OrderService struct {
	*aphgrpc.Service
	repo      repository.OrderRepository
	publisher message.Publisher
}

func defaultOptions() *aphgrpc.ServiceOptions {
	return &aphgrpc.ServiceOptions{Resource: "order"}
}

// NewOrderService is the constructor for creating a new instance of OrderService
func NewOrderService(repo repository.OrderRepository, pub message.Publisher, opt ...aphgrpc.Option) *OrderService {
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

// GetOrder handles getting an order by ID
func (s *OrderService) GetOrder(ctx context.Context, r *order.OrderId) (*order.Order, error) {
	o := &order.Order{}
	if err := r.Validate(); err != nil {
		return o, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	m, err := s.repo.GetOrder(r.Id)
	if err != nil {
		return o, aphgrpc.HandleGetError(ctx, err)
	}
	if m.NotFound {
		return o, aphgrpc.HandleNotFoundError(ctx, err)
	}
	o.Data = &order.Order_Data{
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
	return o, nil
}

// CreateOrder handles the creation of a new order
func (s *OrderService) CreateOrder(ctx context.Context, r *order.NewOrder) (*order.Order, error) {
	o := &order.Order{}
	if err := r.Validate(); err != nil {
		return o, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	m, err := s.repo.AddOrder(r)
	if err != nil {
		return o, aphgrpc.HandleInsertError(ctx, err)
	}
	if m.NotFound {
		return o, aphgrpc.HandleNotFoundError(ctx, err)
	}
	o.Data = &order.Order_Data{
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
	s.publisher.Publish(s.Topics["orderCreate"], o)
	return o, nil
}

// UpdateOrder handles updating an existing order
func (s *OrderService) UpdateOrder(ctx context.Context, r *order.OrderUpdate) (*order.Order, error) {
	o := &order.Order{}
	if err := r.Validate(); err != nil {
		return o, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	m, err := s.repo.EditOrder(r)
	if err != nil {
		return o, aphgrpc.HandleUpdateError(ctx, err)
	}
	if m.NotFound {
		return o, aphgrpc.HandleNotFoundError(ctx, err)
	}
	o.Data = &order.Order_Data{
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
	s.publisher.Publish(s.Topics["orderUpdate"], o)
	return o, nil
}

// ListOrders lists all existing orders
func (s *OrderService) ListOrders(ctx context.Context, r *order.ListParameters) (*order.OrderCollection, error) {
	oc := &order.OrderCollection{}
	var l int64
	c := r.Cursor
	f := r.Filter
	if r.Limit == 0 {
		l = 10
	} else {
		l = r.Limit
	}
	if len(f) > 0 {
		p, err := query.ParseFilterString(f)
		if err != nil {
			return oc, fmt.Errorf("error parsing filter string: %s", err)
		}
		str, err := query.GenAQLFilterStatement(&query.StatementParameters{Fmap: arangodb.FMap, Filters: p, Doc: "s"})
		if err != nil {
			return oc, fmt.Errorf("error generating AQL filter statement: %s", err)
		}
		// if the parsed statement is empty FILTER, just return empty string
		if str == "FILTER " {
			str = ""
		}
		mc, err := s.repo.ListOrders(&order.ListParameters{Cursor: c, Limit: l, Filter: str})
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
func (s *OrderService) LoadOrder(ctx context.Context, r *order.ExistingOrder) (*order.Order, error) {
	o := &order.Order{}
	if err := r.Validate(); err != nil {
		return o, aphgrpc.HandleInvalidParamError(ctx, err)
	}
	m, err := s.repo.LoadOrder(r)
	if err != nil {
		return o, aphgrpc.HandleInsertError(ctx, err)
	}
	if m.NotFound {
		return o, aphgrpc.HandleNotFoundError(ctx, err)
	}
	o.Data = &order.Order_Data{
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
	s.publisher.Publish(s.Topics["orderCreate"], o)
	return o, nil
}

// PrepareForOrder clears the database to prepare for loading data
func (s *OrderService) PrepareForOrder(ctx context.Context, r *empty.Empty) (*empty.Empty, error) {
	e := &empty.Empty{}
	if err := s.repo.ClearOrders(); err != nil {
		return e, aphgrpc.HandleGenericError(ctx, err)
	}
	return e, nil
}

// genNextCursorVal converts to epoch(https://en.wikipedia.org/wiki/Unix_time)
// in milliseconds
func genNextCursorVal(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

func statusToEnum(status string) order.OrderStatus {
	switch status {
	case "Shipped":
		return order.OrderStatus_Shipped
	case "Cancelled":
		return order.OrderStatus_Cancelled
	case "Growing":
		return order.OrderStatus_Growing
	default:
		break
	}
	return order.OrderStatus_In_preparation
}
