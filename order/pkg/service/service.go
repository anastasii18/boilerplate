package service

import (
	"context"
	inventoryModel "inventory/pkg/model"
	"order/pkg/model"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *model.Order, newParts []*inventoryModel.Part) error
	GetOrder(orderUuid string) (*model.Order, error)
	UpdateOrder(orderUuid string, transactionUuid *string, status *model.OrderStatus, paymentMethod *model.OrderPaymentMethod) error
}
