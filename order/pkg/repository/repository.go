package repository

import "order/pkg/model"

type OrderRepository interface {
	CreateOrder(order *Order)
	GetOrder(orderUuid string) (*model.Order, error)
	UpdateOrder(orderUuid string, transactionUuid *string, status *model.OrderStatus, paymentMethod *model.OrderPaymentMethod) error
}
