package service

import (
	repomodel "order/pkg/db"

	"github.com/samber/lo"
)

func OrderToRepoModel(order *Order) *repomodel.Order {
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

func RepoModelToOrder(order *repomodel.Order) *Order {
	return &Order{
		OrderUuid:       order.OrderUuid,
		UserUuid:        order.UserUuid,
		PartUuids:       order.PartUuids,
		TotalPrice:      order.TotalPrice,
		TransactionUuid: order.TransactionUuid,
		PaymentMethod:   OrderPaymentMethod(order.PaymentMethod),
		Status:          OrderStatus(order.Status),
	}
}

func StatusToRepoStatus(status *OrderStatus) *repomodel.OrderStatus {
	return lo.ToPtr(repomodel.OrderStatus(lo.FromPtr(status)))
}

func PaymentMethodToRepoPaymentMethod(paymentMethod *OrderPaymentMethod) *repomodel.OrderPaymentMethod {
	return lo.ToPtr(repomodel.OrderPaymentMethod(lo.FromPtr(paymentMethod)))
}
