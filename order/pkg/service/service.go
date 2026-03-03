package service

import (
	"context"
	"fmt"
	inventoryModel "inventory/pkg/service"
	"order/pkg/db"
	"time"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *Order, newParts []*inventoryModel.Part) error
	GetOrder(orderUuid string) (*Order, error)
	UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

var _ OrderService = (*Service)(nil)

type Service struct {
	OrderRepository db.OrderRepository
}

func NewService(orderRepository db.OrderRepository) *Service {
	return &Service{OrderRepository: orderRepository}
}

func (s Service) CreateOrder(ctx context.Context, order *Order, parts []*inventoryModel.Part) error {
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
	if order.PartUuids == nil {
		order.PartUuids = []string{}
	}

	order.CreatedAt = time.Now()
	order.TotalPrice = totalPrice
	order.OrderUuid = uuid.New().String()
	order.Status = PENDING_PAYMENT

	s.OrderRepository.CreateOrder(OrderToRepoModel(order))

	return nil
}

func (s Service) GetOrder(orderUuid string) (*Order, error) {
	order, err := s.OrderRepository.GetOrder(orderUuid)
	if err != nil {
		return nil, err
	}
	return RepoModelToOrder(order), nil
}

func (s Service) UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error {
	return s.OrderRepository.UpdateOrder(orderUuid, transactionUuid, StatusToRepoStatus(status), PaymentMethodToRepoPaymentMethod(paymentMethod))
}
