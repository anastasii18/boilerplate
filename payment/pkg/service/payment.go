package service

import (
	"context"
	"log"
	"payment/pkg/repository"
	paymentV1 "shared/pkg/proto/payment/v1"
)

var _ PaymentService = (*Service)(nil)

type Service struct {
	paymentRepository repository.PaymentRepository
}

func NewService(paymentRepository repository.PaymentRepository) *Service {
	return &Service{
		paymentRepository: paymentRepository,
	}
}

func (service *Service) PayOrder(context.Context, *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	transactionUuid := service.paymentRepository.PayOrder()
	log.Printf("Оплата прошла успешно, transaction_uuid: %s\n", transactionUuid)
	return &paymentV1.PayOrderResponse{TransactionUuid: transactionUuid}, nil
}
