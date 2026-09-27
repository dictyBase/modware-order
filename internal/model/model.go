// Package model defines data structures for order management.
package model

import (
	"time"

	driver "github.com/arangodb/go-driver"
)

// UserInfo is the compact user profile embedded in an order. It carries
// the mailing and contact fields needed to generate invoices and
// notifications without a lookup against the user service.
type UserInfo struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Organization  string `json:"organization"`
	FirstAddress  string `json:"first_address"`
	SecondAddress string `json:"second_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zipcode       string `json:"zipcode"`
	Country       string `json:"country"`
	Phone         string `json:"phone"`
}

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
	ConsumerInfo     *UserInfo `json:"consumer_info,omitempty"`
	PayerInfo        *UserInfo `json:"payer_info,omitempty"`
	NotFound         bool
}
