package consumer

import (
	"encoding/json"
	"fmt"
	"order/pkg/service"
)

type decoder struct{}

type OrderPaidDecoder interface {
	Decode(data []byte) (service.OrderPaid, error)
}

func NewOrderPaidRecordedDecoder() *decoder {
	return &decoder{}
}

func (d *decoder) Decode(data []byte) (service.OrderPaid, error) {
	var orderPaid service.OrderPaid
	if err := json.Unmarshal(data, &orderPaid); err != nil {
		return service.OrderPaid{}, fmt.Errorf("failed to unmarshal orderPaid: %w", err)
	}

	return service.OrderPaid{
		EventUuid:       orderPaid.EventUuid,
		UserUuid:        orderPaid.UserUuid,
		OrderUuid:       orderPaid.OrderUuid,
		TransactionUuid: orderPaid.TransactionUuid,
		PaymentMethod:   orderPaid.PaymentMethod,
	}, nil
}
