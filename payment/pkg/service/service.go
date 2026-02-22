package service

import (
	"context"
	"log"
	"payment/pkg/repository"
)

type PaymentService interface {
	PayOrder(context.Context) string
}

var _ PaymentService = (*Service)(nil)

type Service struct {
	paymentRepository repository.PaymentRepository
}

func NewService(paymentRepository repository.PaymentRepository) *Service {
	return &Service{
		paymentRepository: paymentRepository,
	}
}

func (service *Service) PayOrder(context.Context) string {
	transactionUuid := service.paymentRepository.PayOrder()
	log.Printf("Оплата прошла успешно, transaction_uuid: %s\n", transactionUuid)
	return transactionUuid
}
