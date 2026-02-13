package repository

import (
	"fmt"
	"order/pkg/model"
	"sync"
)

type Repository struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

var _ OrderRepository = (*Repository)(nil)

func NewRepository() *Repository {
	return &Repository{
		orders: make(map[string]*Order),
	}
}

func (r *Repository) CreateOrder(order *Order) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.OrderUuid] = order
}

func (r *Repository) GetOrder(orderUuid string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[orderUuid]
	if !ok {
		return nil, fmt.Errorf("order with id %s not found", orderUuid)
	}

	return OrderToModel(order), nil
}

func (r *Repository) UpdateOrder(orderUuid string, transactionUuid *string, status *model.OrderStatus, paymentMethod *model.OrderPaymentMethod) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.orders[orderUuid]

	if !ok {
		return fmt.Errorf("order with id %s not found", orderUuid)
	}

	if transactionUuid != nil {
		r.orders[orderUuid].TransactionUuid = *transactionUuid
	}

	if status != nil {
		r.orders[orderUuid].Status = OrderStatus(*status)
	}

	if paymentMethod != nil {
		r.orders[orderUuid].PaymentMethod = OrderPaymentMethod(*paymentMethod)
	}

	return nil
}
