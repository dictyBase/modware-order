package message

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

// Publisher manages publishing of message.
type Publisher interface {
	// Publish publishes the order object using the given subject
	Publish(subject string, ord *order.Order) error
	// Close closes the connection to the underlying messaging server
	Close() error
}
