package service

import (
	"github.com/dictyBase/apihelpers/aphgrpc"
	"github.com/dictyBase/modware-order/internal/message"
	"github.com/dictyBase/modware-order/internal/repository"
)

// OrderService is the container for managing order service
// definition
type OrderService struct {
	*aphgrpc.Service
	repo      repository.OrderRepository
	publisher message.Publisher
}

func defaultOptions() *aphgrpc.ServiceOptions {
	return &aphgrpc.ServiceOptions{Resource: "order"}
}

// NewOrderService is the constructor for creating a new instance of OrderService
func NewOrderService(repo repository.OrderRepository, pub message.Publisher, opt ...aphgrpc.Option) *OrderService {
	so := defaultOptions()
	for _, optfn := range opt {
		optfn(so)
	}
	srv := &aphgrpc.Service{}
	aphgrpc.AssignFieldsToStructs(so, srv)
	return &OrderService{
		Service:   srv,
		repo:      repo,
		publisher: pub,
	}
}
