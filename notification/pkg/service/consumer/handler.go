package consumer

import (
	"context"
	"fmt"
	"platform/pkg/kafka"
	"platform/pkg/logger"

	"go.uber.org/zap"
)

var (
	ShipAssembledTopic       = "ship.assembled"
	OrderPaidTopic           = "order.paid"
	ChatID             int64 = 1 //поменять на действительный
)

func (s *service) OrderHandler(ctx context.Context, msg kafka.Message) error {
	if msg.Topic == ShipAssembledTopic {
		return s.shipAssembledProcess(ctx, msg)
	}
	if msg.Topic == OrderPaidTopic {
		return s.orderPaidProcess(ctx, msg)
	}

	return fmt.Errorf("failed to process message: %s", msg.Topic)
}

func (s *service) orderPaidProcess(ctx context.Context, msg kafka.Message) error {
	event, err := s.orderPaidDecoder.Decode(msg.Value)
	if err != nil {
		logger.Error(ctx, "Failed to decode orderPaidDecoder", zap.Error(err))
		return err
	}

	logger.Info(ctx, "Processing message",
		zap.String("topic", msg.Topic),
		zap.Any("partition", msg.Partition),
		zap.Any("offset", msg.Offset),
		zap.String("order_uuid", event.OrderUuid),
		zap.String("event_uuid", event.EventUuid),
		zap.String("user_uuid", event.UserUuid),
	)
	err = s.telegramService.SendOrderPaidNotification(ctx, event, ChatID)
	if err != nil {
		logger.Error(ctx, "Failed to send order paid notification", zap.Error(err))
		return err
	}

	return nil
}

func (s *service) shipAssembledProcess(ctx context.Context, msg kafka.Message) error {
	event, err := s.shipAssembledDecoder.Decode(msg.Value)
	if err != nil {
		logger.Error(ctx, "Failed to decode shipAssembledDecoder", zap.Error(err))
		return err
	}

	logger.Info(ctx, "Processing message",
		zap.String("topic", msg.Topic),
		zap.Any("partition", msg.Partition),
		zap.Any("offset", msg.Offset),
		zap.String("order_uuid", event.OrderUuid),
		zap.String("event_uuid", event.EventUuid),
		zap.String("user_uuid", event.UserUuid),
	)

	err = s.telegramService.SendAssembledNotification(ctx, event, ChatID)
	if err != nil {
		logger.Error(ctx, "Failed to send order paid notification", zap.Error(err))
		return err
	}

	return nil
}
