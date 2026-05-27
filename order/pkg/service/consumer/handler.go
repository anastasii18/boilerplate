package consumer

import (
	"context"
	orderService "order/pkg/service"
	"platform/pkg/kafka"
	"platform/pkg/logger"

	"go.uber.org/zap"
)

func (s *service) OrderHandler(ctx context.Context, msg kafka.Message) error {
	event, err := s.shipAssembledDecoder.Decode(msg.Value)
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

	err = s.orderService.UpdateOrder(ctx, event.OrderUuid, nil, orderService.Ptr(orderService.COMPLETED), nil)

	if err != nil {
		logger.Error(ctx, "Failed to update order status", zap.Error(err))
		return err
	}

	return nil
}
