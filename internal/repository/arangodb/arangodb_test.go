package arangodb

import (
	"log"
	"os"
	"testing"

	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/arangomanager/testarango"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

var ta *testarango.TestArango

func getConnectParams() *manager.ConnectParams {
	return &manager.ConnectParams{
		User:     ta.User,
		Pass:     ta.Pass,
		Database: ta.Database,
		Host:     ta.Host,
		Port:     ta.Port,
		Istls:    false,
	}
}

func newOrder() *order.NewOrder {
	return &order.NewOrder{
		Data: &order.NewOrder_Data{
			Type: "order",
			Attributes: &order.NewOrderAttributes{
				Courier:  "FedEx",
				Comments: "This is a test comment",
				Payment:  "Credit card",
			},
		},
	}
}

func TestMain(m *testing.M) {
	ta, err := testarango.NewTestArangoFromEnv(true)
	if err != nil {
		log.Fatal(err)
	}
	// clean up the database at the end
	defer func() {
		dbh, err := ta.DB(ta.Database)
		if err != nil {
			log.Fatal(err)
		}
		dbh.Drop()
	}()
	code := m.Run()
	os.Exit(code)
}

// func TestGetOrder(t *testing.T) {
// 	connP := getConnectParams()
// 	_, err := NewOrderRepo(connP, "stock_order")
// 	if err != nil {
// 		t.Fatalf("error in connecting to data source %s", err)
// 	}
// }

// func TestAddOrder(t *testing.T) {

// }

// func TestEditOrder(t *testing.T) {

// }

// func TestListOrders(t *testing.T) {

// }
