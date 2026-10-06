// Package repository defines the interface for order data access.
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
	// Autocomplete suggests orders for a partial search text. It runs
	// exact prefix matching first, then fills the remaining slots with
	// fuzzy matches. A query without matches returns an empty slice.
	Autocomplete(query string, limit int) ([]*Suggestion, error)
}

// Suggestion is a single autocomplete match for an order.
type Suggestion struct {
	// ID is the unique identifier of the matched order.
	ID string
	// Field is the name of the order field that matched, for example
	// purchase_order_num or consumer_info.organization.
	Field string
	// DisplayText is the matched value, shown as the suggestion text.
	DisplayText string
	// Score is the search score assigned by the ArangoSearch view.
	Score float64
}
