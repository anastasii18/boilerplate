package service

import (
	repomodel "order/pkg/db"
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
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
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
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func StatusToRepoStatus(status *OrderStatus) *repomodel.OrderStatus {
	return Ptr(repomodel.OrderStatus(Val(status)))
}

func PaymentMethodToRepoPaymentMethod(paymentMethod *OrderPaymentMethod) *repomodel.OrderPaymentMethod {
	return Ptr(repomodel.OrderPaymentMethod(Val(paymentMethod)))
}

func Val[T any, P *T](p P) T {
	if p != nil {
		return *p
	}
	var def T
	return def
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
