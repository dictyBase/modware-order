// Package nats provides NATS-based message publisher implementation.
package nats

import (
	"fmt"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/message"
	gnats "github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type natsPublisher struct {
	conn *gnats.Conn
}

// NewPublisher creates a new NATS publisher instance connected to the specified host and port.
func NewPublisher(
	host, port string,
	options ...gnats.Option,
) (message.Publisher, error) {
	ntc, err := gnats.Connect(
		fmt.Sprintf("nats://%s:%s", host, port),
		options...)
	if err != nil {
		return &natsPublisher{}, fmt.Errorf(
			"unable to connect to nats server %s",
			err,
		)
	}

	return &natsPublisher{conn: ntc}, nil
}

func (n *natsPublisher) Publish(subj string, ord *order.Order) error {
	data, err := proto.Marshal(ord)
	if err != nil {
		return fmt.Errorf("error marshaling order: %s", err)
	}
	err = n.conn.Publish(subj, data)
	if err != nil {
		return fmt.Errorf("error in publishing to nats server %s", err)
	}

	return nil
}

func (n *natsPublisher) Close() error {
	n.conn.Close()

	return nil
}
