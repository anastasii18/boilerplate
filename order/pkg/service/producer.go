package service

import (
	"context"
	"encoding/json"
	"platform/pkg/kafka"
)

type OrderProducerService interface {
	ProduceOrderPaid(ctx context.Context, event OrderPaid) error
}

func NewProducerService(orderPaidProducer kafka.Producer) *producer {
	return &producer{
		orderPaidProducer: orderPaidProducer,
	}
}

type producer struct {
	orderPaidProducer kafka.Producer
}

type OrderPaid struct {
	EventUuid       string  `json:"event_uuid"`       // Уникальный идентификатор события (для идемпотентности)
	OrderUuid       string  `json:"order_uuid"`       // Идентификатор оплаченного заказа
	UserUuid        string  `json:"user_uuid"`        // Идентификатор пользователя
	TransactionUuid *string `json:"transaction_uuid"` // Идентификатор транзакции, сгенерированный в результате оплаты
	PaymentMethod   string  `json:"payment_method"`   // Способ оплаты
}

func (p *producer) ProduceOrderPaid(ctx context.Context, event OrderPaid) error {
	eventBytes, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	return p.orderPaidProducer.Send(ctx, []byte(event.EventUuid), eventBytes)
}
