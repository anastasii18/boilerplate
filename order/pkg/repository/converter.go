package repository

import "order/pkg/model"

func OrderToModel(order *Order) *model.Order {
	return &model.Order{
		OrderUuid:       order.OrderUuid,
		UserUuid:        order.UserUuid,
		PartUuids:       order.PartUuids,
		TotalPrice:      order.TotalPrice,
		TransactionUuid: order.TransactionUuid,
		PaymentMethod:   model.OrderPaymentMethod(order.PaymentMethod),
		Status:          model.OrderStatus(order.Status),
	}
}
