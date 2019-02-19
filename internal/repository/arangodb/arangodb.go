package arangodb

import (
	"context"
	"fmt"
	"strings"

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
	ar := &arangorepository{}
	sess, db, err := manager.NewSessionDb(connP)
	if err != nil {
		return ar, err
	}
	ar.sess = sess
	ar.database = db
	sorderc, err := db.FindOrCreateCollection(coll, &driver.CreateCollectionOptions{})
	if err != nil {
		return ar, err
	}
	ar.sorder = sorderc
	return ar, nil
}

// GetOrder retrieves stock order from database
func (ar *arangorepository) GetOrder(id string) (*model.OrderDoc, error) {
	m := &model.OrderDoc{}
	bindVars := map[string]interface{}{
		"@stock_order_collection": ar.sorder.Name(),
		"key":                     id,
	}
	r, err := ar.database.GetRow(orderGet, bindVars)
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
		"status":                  attr.Status.String(),
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
	m := &model.OrderDoc{}
	attr := uo.Data.Attributes
	// check if order exists
	em, err := ar.GetOrder(uo.Data.Id)
	if err != nil {
		return m, err
	}
	if em.NotFound {
		m.NotFound = true
		return m, nil
	}
	bindVars := getUpdatableBindParams(attr)
	var bindParams []string
	for k := range bindVars {
		bindParams = append(bindParams, fmt.Sprintf("%s: @%s", k, k))
	}
	orderUpdQ := fmt.Sprintf(orderUpd, strings.Join(bindParams, ","))
	bindVars["@stock_order_collection"] = ar.sorder.Name()
	bindVars["key"] = uo.Data.Id

	rupd, err := ar.database.DoRun(orderUpdQ, bindVars)
	if err != nil {
		return m, err
	}
	if err := rupd.Read(m); err != nil {
		return m, err
	}
	return m, nil
}

// ListOrders provides a list of all orders
func (ar *arangorepository) ListOrders(p *order.ListParameters) ([]*model.OrderDoc, error) {
	var om []*model.OrderDoc
	var stmt string
	c := p.Cursor
	l := p.Limit
	f := p.Filter
	if len(f) > 0 {
		if c == 0 { // no cursor so return first set of result
			stmt = fmt.Sprintf(
				orderListWithFilter,
				ar.sorder.Name(),
				f,
				l+1,
			)
		} else { // else include both filter and cursor
			stmt = fmt.Sprintf(
				orderListFilterWithCursor,
				ar.sorder.Name(),
				c,
				f,
				l+1,
			)
		}
	} else {
		if c == 0 {
			stmt = fmt.Sprintf(
				orderList,
				ar.sorder.Name(),
				l+1,
			)
		} else {
			stmt = fmt.Sprintf(
				orderListWithCursor,
				ar.sorder.Name(),
				c,
				l+1,
			)
		}
	}
	rs, err := ar.database.Search(stmt)
	if err != nil {
		return om, err
	}
	if rs.IsEmpty() {
		return om, nil
	}
	for rs.Scan() {
		m := &model.OrderDoc{}
		if err := rs.Read(m); err != nil {
			return om, err
		}
		om = append(om, m)
	}
	return om, nil
}

func getUpdatableBindParams(attr *order.OrderUpdateAttributes) map[string]interface{} {
	bindVars := make(map[string]interface{})
	if len(attr.Courier) > 0 {
		bindVars["courier"] = attr.Courier
	}
	if len(attr.CourierAccount) > 0 {
		bindVars["courier_account"] = attr.CourierAccount
	}
	if len(attr.Comments) > 0 {
		bindVars["comments"] = attr.Comments
	}
	if len(attr.Payment) > 0 {
		bindVars["payment"] = attr.Payment
	}
	if len(attr.PurchaseOrderNum) > 0 {
		bindVars["purchase_order_num"] = attr.PurchaseOrderNum
	}
	bindVars["status"] = attr.Status.String()
	if len(attr.Items) > 0 {
		bindVars["items"] = attr.Items
	}
	return bindVars
}

// ClearOrders clears all orders from the repository datasource
func (ar *arangorepository) ClearOrders() error {
	if err := ar.sorder.Truncate(context.Background()); err != nil {
		return err
	}
	return nil
}
