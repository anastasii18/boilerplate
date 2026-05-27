package consumer

import (
	"assembly/pkg/service/producer"
	"context"
	"platform/pkg/kafka"
	"platform/pkg/kafka/consumer"
	"platform/pkg/logger"

	"go.uber.org/zap"
)

type OrderPaidService interface {
	RunConsumer(ctx context.Context) error
}

var _ OrderPaidService = (*service)(nil)

type service struct {
	orderPaidConsumer kafka.Consumer
	orderPaidDecoder  consumer.OrderPaidDecoder
	shipProducer      producer.ShipAssembledService
}

func NewService(orderPaidRecordedConsumer kafka.Consumer, orderPaidRecordedDecoder consumer.OrderPaidDecoder, shipProducer producer.ShipAssembledService) *service {
	return &service{
		orderPaidConsumer: orderPaidRecordedConsumer,
		orderPaidDecoder:  orderPaidRecordedDecoder,
		shipProducer:      shipProducer,
	}
}

func (s *service) RunConsumer(ctx context.Context) error {
	logger.Info(ctx, "Starting order orderPaidConsumer service")

	err := s.orderPaidConsumer.Consume(ctx, s.OrderHandler)
	if err != nil {
		logger.Error(ctx, "Consume from order.paid topic error", zap.Error(err))
		return err
	}

	return nil
}
