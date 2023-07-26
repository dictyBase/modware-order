package repository

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/model"
)

// OrderRepository is an interface for accessing
// stock order data.
type OrderRepository interface {
	GetOrder(id string) (*model.OrderDoc, error)
	AddOrder(no *order.NewOrder) (*model.OrderDoc, error)
	EditOrder(uo *order.OrderUpdate) (*model.OrderDoc, error)
	ListOrders(p *order.ListParameters) ([]*model.OrderDoc, error)
	LoadOrder(eo *order.ExistingOrder) (*model.OrderDoc, error)
	ClearOrders() error
}
