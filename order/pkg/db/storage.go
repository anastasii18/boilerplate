package db

import (
	"sync"
)

type OrderStorage struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string]*Order),
	}
}

func (storage *OrderStorage) CreateOrder(order *Order) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.orders[order.OrderUuid] = order
}

func (storage *OrderStorage) GetOrder(orderUuid string) *Order {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	return storage.orders[orderUuid]
}
