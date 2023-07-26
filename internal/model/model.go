package model

import (
	"time"

	driver "github.com/arangodb/go-driver"
)

// OrderDoc is the data structure for stock orders.
type OrderDoc struct {
	driver.DocumentMeta
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Courier          string    `json:"courier"`
	CourierAccount   string    `json:"courier_account"`
	Comments         string    `json:"comments"`
	Payment          string    `json:"payment"`
	PurchaseOrderNum string    `json:"purchase_order_num"`
	Status           string    `json:"status"`
	Consumer         string    `json:"consumer"`
	Payer            string    `json:"payer"`
	Purchaser        string    `json:"purchaser"`
	Items            []string  `json:"items"`
	NotFound         bool
}
