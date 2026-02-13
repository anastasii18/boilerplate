package converter

import (
	"order/pkg/model"
	repomodel "order/pkg/repository"
)

func OrderToRepoModel(order *model.Order) *repomodel.Order {
	return &repomodel.Order{
		OrderUuid:       order.OrderUuid,
		UserUuid:        order.UserUuid,
		PartUuids:       order.PartUuids,
		TotalPrice:      order.TotalPrice,
		TransactionUuid: order.TransactionUuid,
		PaymentMethod:   repomodel.OrderPaymentMethod(order.PaymentMethod),
		Status:          repomodel.OrderStatus(order.Status),
	}
}
