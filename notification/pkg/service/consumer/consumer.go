package consumer

import (
	"context"
	notifo "notification/pkg/service"
	"platform/pkg/kafka"
	"platform/pkg/kafka/consumer"
	"platform/pkg/kafka/producer"
	"platform/pkg/logger"

	"go.uber.org/zap"
)

type ShipAndOrderNotificationService interface {
	RunConsumer(ctx context.Context) error
}

var _ ShipAndOrderNotificationService = (*service)(nil)

type service struct {
	notificationConsumer kafka.Consumer
	shipAssembledDecoder producer.ShipAssembledDecoder
	orderPaidDecoder     consumer.OrderPaidDecoder
	telegramService      *notifo.TGService
}

func NewService(notificationConsumer kafka.Consumer, shipAssembledDecoder producer.ShipAssembledDecoder, orderPaidDecoder consumer.OrderPaidDecoder, telegramService *notifo.TGService) *service {
	return &service{
		notificationConsumer: notificationConsumer,
		shipAssembledDecoder: shipAssembledDecoder,
		orderPaidDecoder:     orderPaidDecoder,
		telegramService:      telegramService,
	}
}

func (s *service) RunConsumer(ctx context.Context) error {
	logger.Info(ctx, "Starting notification consumer service")

	err := s.notificationConsumer.Consume(ctx, s.OrderHandler)
	if err != nil {
		logger.Error(ctx, "Consume from ship.assembled or order.paid topic error", zap.Error(err))
		return err
	}

	return nil
}
