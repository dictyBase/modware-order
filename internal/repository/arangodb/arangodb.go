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

func (ar *arangorepository) GetOrder(id string) (*model.OrderDoc, error) {

}

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

func (ar *arangorepository) EditOrder(uo *order.OrderUpdate) (*model.OrderDoc, error) {

}

func (ar *arangorepository) ListOrders(cursor int64, limit int64) ([]*model.OrderDoc, error) {

}
