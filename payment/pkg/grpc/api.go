package grpc

import (
	"context"
	"payment/pkg/service"
	paymentV1 "shared/pkg/proto/payment/v1"
)

// реализует gRPC сервис для работы с транзакциями
type Api struct {
	paymentV1.UnimplementedPaymentServiceServer

	paymentService service.PaymentService
}

func NewApi(paymentService service.PaymentService) *Api {
	return &Api{paymentService: paymentService}
}

func (service *Api) PayOrder(ctx context.Context, request *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	return &paymentV1.PayOrderResponse{TransactionUuid: service.paymentService.PayOrder(ctx)}, nil
}
