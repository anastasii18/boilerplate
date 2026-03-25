package producer

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	orderProducer "order/pkg/service"
	"platform/pkg/kafka"
	"platform/pkg/logger"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ShipAssembled struct {
	EventUuid    string `json:"event_uuid"`     // Уникальный идентификатор события (для идемпотентности)
	OrderUuid    string `json:"order_uuid"`     // Идентификатор собранного заказа
	UserUuid     string `json:"user_uuid"`      // Идентификатор пользователя
	BuildTimeSec int64  `json:"build_time_sec"` // Время (в секундах), потраченное на сборку корабля
}

var _ ShipAssembledService = (*producer)(nil)

type ShipAssembledService interface {
	ProduceOrderPaid(ctx context.Context, event orderProducer.OrderPaid) error
}

func NewService(shipAssembledProducer kafka.Producer) *producer {
	return &producer{
		shipAssembledProducer: shipAssembledProducer,
	}
}

type producer struct {
	shipAssembledProducer kafka.Producer
}

func (p *producer) ProduceOrderPaid(ctx context.Context, event orderProducer.OrderPaid) error {
	start := time.Now()
	// имитация сборки
	simulateAssembly(ctx, event.OrderUuid)
	duration := time.Since(start)

	shipEvent := ShipAssembled{
		EventUuid:    uuid.New().String(),
		OrderUuid:    event.OrderUuid,
		UserUuid:     event.UserUuid,
		BuildTimeSec: int64(duration.Seconds()),
	}
	eventBytes, err := json.Marshal(&shipEvent)
	if err != nil {
		return err
	}

	return p.shipAssembledProducer.Send(ctx, []byte(shipEvent.EventUuid), eventBytes)
}

func simulateAssembly(ctx context.Context, orderUuid string) {
	// Имитация работы сборочного цеха
	n, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		logger.Error(ctx, "failed to generate random number", zap.Error(err))
		n = big.NewInt(5)
	}
	duration := time.Duration(10+n.Int64()) * time.Second
	logger.Info(ctx, "Starting assembly simulation", zap.String("order_uuid", orderUuid), zap.Duration("estimated_duration", duration))

	select {
	case <-time.After(duration):
		logger.Info(ctx, "Assembly simulation completed", zap.String("order_uuid", orderUuid))
	case <-ctx.Done():
		logger.Warn(ctx, "Assembly cancelled", zap.String("order_uuid", orderUuid))
	}
}
