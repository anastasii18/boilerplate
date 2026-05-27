package service

import (
	"context"
	"fmt"
	"order/pkg"
	"order/pkg/client/inventory"
	"order/pkg/db"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, orderUuid string) (*Order, error)
	UpdateOrder(ctx context.Context, orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

var _ OrderService = (*Service)(nil)

type Service struct {
	repo            db.OrderRepository
	inventoryClient *inventory.Client
	producerService OrderProducerService
}

func NewService(orderRepository db.OrderRepository, inventoryClient *inventory.Client, producerService OrderProducerService) *Service {
	return &Service{repo: orderRepository, inventoryClient: inventoryClient, producerService: producerService}
}

func (s Service) CreateOrder(ctx context.Context, order *Order) error {
	ctx = addSessionUuid(ctx)
	parts, err := s.inventoryClient.GetInventoryPartsForIDs(ctx, order.PartUuids)
	if err != nil {
		return err
	}

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

	err = s.repo.CreateOrder(ctx, OrderToRepoModel(order))
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
	err := s.repo.UpdateOrder(ctx, orderUuid, transactionUuid, StatusToRepoStatus(status), PaymentMethodToRepoPaymentMethod(paymentMethod))
	if err != nil {
		return err
	}

	if Val(status) == PAID {
		userUuid, err := s.repo.GetUserUuidForOrder(ctx, orderUuid)
		if err != nil {
			return err
		}

		err = s.producerService.ProduceOrderPaid(ctx, OrderPaid{
			EventUuid:       uuid.New().String(),
			OrderUuid:       orderUuid,
			UserUuid:        Val(userUuid),
			TransactionUuid: transactionUuid,
			PaymentMethod:   paymentMethod.String(),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func addSessionUuid(ctx context.Context) context.Context {
	val := ctx.Value(pkg.SessionUUIDKey)

	if str, ok := val.(string); ok && str != "" {
		return metadata.AppendToOutgoingContext(ctx, "session-uuid", str)
	}

	return ctx
}
