package nats

import (
	"fmt"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/message"
	gnats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
)

type natsPublisher struct {
	econn *gnats.EncodedConn
}

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
	enc, err := gnats.NewEncodedConn(ntc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return &natsPublisher{}, fmt.Errorf(
			"error in connecting to nats server %s",
			err,
		)
	}

	return &natsPublisher{econn: enc}, nil
}

func (n *natsPublisher) Publish(subj string, ord *order.Order) error {
	err := n.econn.Publish(subj, ord)
	if err != nil {
		return fmt.Errorf("error in publishing to nats server %s", err)
	}

	return nil
}

func (n *natsPublisher) Close() error {
	n.econn.Close()

	return nil
}
