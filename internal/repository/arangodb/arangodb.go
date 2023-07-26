package arangodb

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// NewOrderRepo acts as constructor for database.
func NewOrderRepo(
	connP *manager.ConnectParams,
	coll string,
) (repository.OrderRepository, error) {
	arp := &arangorepository{}
	sess, dbs, err := manager.NewSessionDb(connP)
	if err != nil {
		return arp, fmt.Errorf("error in getting new session %s", err)
	}
	arp.sess = sess
	arp.database = dbs
	sorderc, err := dbs.FindOrCreateCollection(
		coll,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return arp, fmt.Errorf(
			"error in finding or creating collection %s",
			err,
		)
	}
	arp.sorder = sorderc

	return arp, nil
}

// GetOrder retrieves stock order from database.
func (ar *arangorepository) GetOrder(id string) (*model.OrderDoc, error) {
	mdl := &model.OrderDoc{}
	bindVars := map[string]interface{}{
		"@stock_order_collection": ar.sorder.Name(),
		"key":                     id,
	}
	dbr, err := ar.database.GetRow(orderGet, bindVars)
	if err != nil {
		return mdl, fmt.Errorf("error in getting database row %s", err)
	}
	if dbr.IsEmpty() {
		mdl.NotFound = true

		return mdl, nil
	}
	if err := dbr.Read(mdl); err != nil {
		return mdl, fmt.Errorf("error in reading struct %s", err)
	}

	return mdl, nil
}

// AddOrder creates a new stock order.
func (ar *arangorepository) AddOrder(
	no *order.NewOrder,
) (*model.OrderDoc, error) {
	mdl := &model.OrderDoc{}
	var bindVars map[string]interface{}
	attr := no.Data.Attributes
	bindVars = addableOrderBindParams(attr)
	bindVars["@stock_order_collection"] = ar.sorder.Name()
	r, err := ar.database.DoRun(orderIns, bindVars)
	if err != nil {
		return mdl, fmt.Errorf("error in running database command %s", err)
	}
	if err := r.Read(mdl); err != nil {
		return mdl, fmt.Errorf("error in reading struct %s", err)
	}

	return mdl, nil
}

// EditOrder updates an existing order.
func (ar *arangorepository) EditOrder(
	uod *order.OrderUpdate,
) (*model.OrderDoc, error) {
	mdl := &model.OrderDoc{}
	attr := uod.Data.Attributes
	// check if order exists
	em, err := ar.GetOrder(uod.Data.Id)
	if err != nil {
		return mdl, err
	}
	if em.NotFound {
		mdl.NotFound = true

		return mdl, nil
	}
	bindVars := getUpdatableBindParams(attr)
	bindParams := make([]string, 0)
	for k := range bindVars {
		bindParams = append(bindParams, fmt.Sprintf("%s: @%s", k, k))
	}
	orderUpdQ := fmt.Sprintf(orderUpd, strings.Join(bindParams, ","))
	bindVars["@stock_order_collection"] = ar.sorder.Name()
	bindVars["key"] = uod.Data.Id

	rupd, err := ar.database.DoRun(orderUpdQ, bindVars)
	if err != nil {
		return mdl, fmt.Errorf("error in running statement in database %s", err)
	}
	if err := rupd.Read(mdl); err != nil {
		return mdl, fmt.Errorf(
			"error in binding struct from database query %s",
			err,
		)
	}

	return mdl, nil
}

func (ar *arangorepository) stmtWithFilter(pls *order.ListParameters) string {
	if pls.Cursor == 0 { // no cursor so return first set of result
		return fmt.Sprintf(
			orderListWithFilter,
			ar.sorder.Name(),
			pls.Filter,
			pls.Limit+1,
		)
	}

	return fmt.Sprintf(
		orderListFilterWithCursor,
		ar.sorder.Name(),
		pls.Cursor,
		pls.Filter,
		pls.Limit+1,
	)
}

func (ar *arangorepository) stmtWithOutFilter(
	pls *order.ListParameters,
) string {
	if pls.Cursor == 0 {
		return fmt.Sprintf(
			orderList,
			ar.sorder.Name(),
			pls.Limit+1,
		)
	}

	return fmt.Sprintf(
		orderListWithCursor,
		ar.sorder.Name(),
		pls.Cursor,
		pls.Limit+1,
	)
}

// ListOrders provides a list of all orders.
func (ar *arangorepository) ListOrders(
	pls *order.ListParameters,
) ([]*model.OrderDoc, error) {
	var odr []*model.OrderDoc
	var stmt string
	if len(pls.Filter) > 0 {
		stmt = ar.stmtWithFilter(pls)
	} else {
		stmt = ar.stmtWithOutFilter(pls)
	}
	rsd, err := ar.database.Search(stmt)
	if err != nil {
		return odr, fmt.Errorf("error in database searching %s", err)
	}
	if rsd.IsEmpty() {
		return odr, nil
	}
	for rsd.Scan() {
		m := &model.OrderDoc{}
		if err := rsd.Read(m); err != nil {
			return odr, fmt.Errorf("error in binding struct %s", err)
		}
		odr = append(odr, m)
	}

	return odr, nil
}

func (ar *arangorepository) LoadOrder(
	eo *order.ExistingOrder,
) (*model.OrderDoc, error) {
	mrd := &model.OrderDoc{}
	bindVars := existingOrderBindParams(eo.Data.Attributes)
	bindVars["@stock_order_collection"] = ar.sorder.Name()
	r, err := ar.database.DoRun(orderLoad, bindVars)
	if err != nil {
		return mrd, fmt.Errorf("error in running database query %s", err)
	}
	if err := r.Read(mrd); err != nil {
		return mrd, fmt.Errorf("error in binding to struct %s", err)
	}

	return mrd, nil
}

func getUpdatableBindParams(
	attr *order.OrderUpdateAttributes,
) map[string]interface{} {
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

// ClearOrders clears all orders from the repository datasource.
func (ar *arangorepository) ClearOrders() error {
	if err := ar.sorder.Truncate(context.Background()); err != nil {
		return fmt.Errorf("error in truncating %s", err)
	}

	return nil
}

func addableOrderBindParams(
	attr *order.NewOrderAttributes,
) map[string]interface{} {
	return map[string]interface{}{
		"courier":            attr.Courier,
		"courier_account":    attr.CourierAccount,
		"comments":           normalizeStrBindParam(attr.Comments),
		"payment":            attr.Payment,
		"purchase_order_num": attr.PurchaseOrderNum,
		"status":             attr.Status.String(),
		"consumer":           attr.Consumer,
		"payer":              attr.Payer,
		"purchaser":          attr.Purchaser,
		"items":              attr.Items,
	}
}

func existingOrderBindParams(
	attr *order.ExistingOrderAttributes,
) map[string]interface{} {
	ctime := attr.CreatedAt.AsTime().Format(time.RFC3339)

	return map[string]interface{}{
		"created_at": ctime,
		"updated_at": ctime,
		"purchaser":  attr.Purchaser,
		"items":      attr.Items,
	}
}

func normalizeStrBindParam(str string) string {
	if len(str) > 0 {
		return str
	}

	return ""
}
