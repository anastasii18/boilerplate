package consumer

import (
	"context"
	"platform/pkg/kafka"
	"platform/pkg/kafka/producer"
	"platform/pkg/logger"

	orderService "order/pkg/service"

	"go.uber.org/zap"
)

type ShipAssembledService interface {
	RunConsumer(ctx context.Context) error
}

var _ ShipAssembledService = (*service)(nil)

type service struct {
	shipAssembledConsumer kafka.Consumer
	shipAssembledDecoder  producer.ShipAssembledDecoder
	orderService          orderService.OrderService
}

func NewService(shipAssembledConsumer kafka.Consumer, shipAssembledDecoder producer.ShipAssembledDecoder, orderService orderService.OrderService) *service {
	return &service{
		shipAssembledConsumer: shipAssembledConsumer,
		shipAssembledDecoder:  shipAssembledDecoder,
		orderService:          orderService,
	}
}

func (s *service) RunConsumer(ctx context.Context) error {
	logger.Info(ctx, "Starting order shipAssembledConsumer service")

	err := s.shipAssembledConsumer.Consume(ctx, s.OrderHandler)
	if err != nil {
		logger.Error(ctx, "Consume from ship.assembled topic error", zap.Error(err))
		return err
	}

	return nil
}
