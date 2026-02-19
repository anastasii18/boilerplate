package db

import (
	"fmt"
	"sync"
)

type OrderRepository interface {
	CreateOrder(order *Order)
	GetOrder(orderUuid string) (*Order, error)
	UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

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

func (r *Repository) GetOrder(orderUuid string) (*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.orders[orderUuid]
	if !ok {
		return nil, fmt.Errorf("order with id %s not found", orderUuid)
	}

	return order, nil
}

func (r *Repository) UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error {
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
		r.orders[orderUuid].Status = *status
	}

	if paymentMethod != nil {
		r.orders[orderUuid].PaymentMethod = *paymentMethod
	}

	return nil
}
