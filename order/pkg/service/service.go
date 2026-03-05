package service

import (
	"context"
	"fmt"
	"order/pkg/client/inventory"
	"order/pkg/db"
	"time"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, orderUuid string) (*Order, error)
	UpdateOrder(ctx context.Context, orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

var _ OrderService = (*Service)(nil)

type Service struct {
	repo            db.OrderRepository
	inventoryClient inventory.Client
}

func NewService(orderRepository db.OrderRepository, inventoryClient inventory.Client) *Service {
	return &Service{repo: orderRepository, inventoryClient: inventoryClient}
}

func (s Service) CreateOrder(ctx context.Context, order *Order) error {
	parts, _ := s.inventoryClient.GetInventoryPartsForIDs(ctx, order.PartUuids)

	// Проверяет, что все детали существуют. Если хотя бы одной нет — возвращает ошибку
	if len(order.PartUuids) != len(parts) {
		return fmt.Errorf("one of part not found")
	}
	totalPrice := 0.0
	for _, part := range parts {
		if part.StockQuantity < 1 {
			return fmt.Errorf("one of part stock quantity not found")
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

	err := s.repo.CreateOrder(ctx, OrderToRepoModel(order))
	if err != nil {
		return err
	}

	return nil
}

func (s Service) GetOrder(ctx context.Context, orderUuid string) (*Order, error) {
	order, err := s.repo.GetOrder(ctx, orderUuid)
	if err != nil {
		return nil, err
	}
	return RepoModelToOrder(order), nil
}

func (s Service) UpdateOrder(ctx context.Context, orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error {
	return s.repo.UpdateOrder(ctx, orderUuid, transactionUuid, StatusToRepoStatus(status), PaymentMethodToRepoPaymentMethod(paymentMethod))
}
