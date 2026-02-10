package grpc

import (
	"context"
	"log"
	paymentV1 "shared/pkg/proto/payment/v1"

	"github.com/google/uuid"
)

// реализует gRPC сервис для работы с транзакциями
type PaymentService struct {
	paymentV1.UnimplementedPaymentServiceServer
}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (service *PaymentService) PayOrder(context.Context, *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	transactionUuid := uuid.New().String()
	log.Printf("Оплата прошла успешно, transaction_uuid: %s\n", transactionUuid)
	return &paymentV1.PayOrderResponse{TransactionUuid: transactionUuid}, nil
}
