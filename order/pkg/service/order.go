package service

import (
	"context"
	"fmt"
	inventoryModel "inventory/pkg/model"
	"order/pkg/converter"
	"order/pkg/model"
	"order/pkg/repository"

	"github.com/google/uuid"
)

var _ OrderService = (*Service)(nil)

type Service struct {
	OrderRepository repository.OrderRepository
}

func NewService() *Service {
	return &Service{OrderRepository: repository.NewRepository()}
}

func (s Service) CreateOrder(ctx context.Context, order *model.Order, parts []*inventoryModel.Part) error {
	// Проверяет, что все детали существуют. Если хотя бы одной нет — возвращает ошибку
	if len(order.PartUuids) != len(parts) {
		return fmt.Errorf("One of part not found")
	}
	totalPrice := 0.0
	for _, part := range parts {
		if part.StockQuantity < 1 {
			return fmt.Errorf("One of part stock quantity not found")
		}
		totalPrice += part.Price
	}

	order.TotalPrice = totalPrice
	order.OrderUuid = uuid.New().String()
	order.Status = model.PENDING_PAYMENT

	s.OrderRepository.CreateOrder(converter.OrderToRepoModel(order))

	return nil
}

func (s Service) GetOrder(orderUuid string) (*model.Order, error) {
	return s.OrderRepository.GetOrder(orderUuid)
}

func (s Service) UpdateOrder(orderUuid string, transactionUuid *string, status *model.OrderStatus, paymentMethod *model.OrderPaymentMethod) error {
	return s.OrderRepository.UpdateOrder(orderUuid, transactionUuid, status, paymentMethod)
}
