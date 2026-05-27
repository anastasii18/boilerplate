package consumer

import (
	"context"
	"platform/pkg/kafka"
	"platform/pkg/logger"
	"time"

	"go.uber.org/zap"
)

func (s *service) OrderHandler(ctx context.Context, msg kafka.Message) error {
	event, err := s.orderPaidDecoder.Decode(msg.Value)
	if err != nil {
		logger.Error(ctx, "Failed to decode orderPaidDecoder", zap.Error(err))
		return err
	}

	if err := sleepWithContext(ctx, 10*time.Second); err != nil {
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

	if err := s.shipProducer.ProduceOrderPaid(ctx, event); err != nil {
		logger.Error(ctx, "Failed to send ShipAssembled", zap.Error(err))
		return err
	}

	return nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
