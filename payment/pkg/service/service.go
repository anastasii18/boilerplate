package service

import (
	"context"
	paymentV1 "shared/pkg/proto/payment/v1"
)

type PaymentService interface {
	PayOrder(context.Context, *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error)
}
