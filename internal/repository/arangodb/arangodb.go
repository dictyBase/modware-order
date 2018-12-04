package arangodb

import (
	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/model"
	"github.com/dictyBase/modware-order/internal/repository"
)

type arangorepository struct {
	sess     *manager.Session
	database *manager.Database
	sorder   driver.Collection
}

// NewOrderRepo acts as constructor for database
func NewOrderRepo(connP *manager.ConnectParams, coll string) (repository.OrderRepository, error) {
	var ar *arangorepository
	sess, db, err := manager.NewSessionDb(connP)
	if err != nil {
		return ar, err
	}
	ar.sess = sess
	ar.database = db
	sorderc, err = db.Collection(coll)
	if err != nil {
		return ar, err
	}
	ar.sorder = sorderc
	return ar, nil
}

// GetOrder retrieves stock order from database
func (ar *arangorepository) GetOrder(id string) (*model.OrderDoc, error) {
	m := &model.OrderDoc{}
	r, err := ar.database.Get(orderGet)
	if err != nil {
		return m, err
	}
	if r.isEmpty() {
		m.NotFound = true
		return m, nil
	}
	if err := r.Read(m); err != nil {
		return m, err
	}
	return m, nil
}

// AddOrder creates a new stock order
func (ar *arangorepository) AddOrder(no *order.NewOrder) (*model.OrderDoc, error) {
	m := &model.OrderDoc{}
	attr := no.Data.Attributes
	bindVars := map[string]interface{}{
		"@stock_order_collection": ar.sorder.Name(),
		"courier":                 attr.Courier,
		"courier_account":         attr.CourierAccount,
		"comments":                attr.Comments,
		"payment":                 attr.Payment,
		"purchase_order_num":      attr.PurchaseOrderNum,
		"status":                  attr.Status,
		"consumer":                attr.Consumer,
		"payer":                   attr.Payer,
		"purchaser":               attr.Purchaser,
		"items":                   attr.Items,
	}
	r, err := ar.database.DoRun(orderIns, bindVars)
	if err != nil {
		return m, err
	}
	if err := r.Read(m); err != nil {
		return m, err
	}
	return m, nil
}

// EditOrder updates an existing order
func (ar *arangorepository) EditOrder(uo *order.OrderUpdate) (*model.OrderDoc, error) {
	m := &model.AnnoDoc{}
	attr := uo.Data.Attributes
	// check if order exists
	r, err := ar.database.Get(orderGet)
	if err != nil {
		return m, err
	}
	if r.IsEmpty() {
		m.NotFound = true
		return m, nil
	}
	if err := r.Read(m); err != nil {
		return m, err
	}
	bindVars := map[string]interface{}{
		"@stock_order_collection": ar.sorder.Name(),
		"courier":                 attr.Courier,
		"courier_account":         attr.CourierAccount,
		"comments":                attr.Comments,
		"payment":                 attr.Payment,
		"purchase_order_num":      attr.PurchaseOrderNum,
		"status":                  attr.Status,
		"items":                   attr.Items,
	}
	rupd, err := ar.database.DoRun(orderUpd, bindVars)
	if err != nil {
		return m, err
	}
	if err := rupd.Read(m); err != nil {
		return m, err
	}
	return m, nil
}

// ListOrders provides a list of all orders
func (ar *arangorepository) ListOrders(cursor int64, limit int64) ([]*model.OrderDoc, error) {

}
