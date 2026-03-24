package producer

import (
	assemblyProducer "assembly/pkg/service/producer"
	"encoding/json"
	"fmt"
)

type decoder struct{}

type ShipAssembledDecoder interface {
	Decode(data []byte) (assemblyProducer.ShipAssembled, error)
}

func NewShipAssembledDecoder() *decoder {
	return &decoder{}
}

func (d *decoder) Decode(data []byte) (assemblyProducer.ShipAssembled, error) {
	var shipAssembled assemblyProducer.ShipAssembled
	if err := json.Unmarshal(data, &shipAssembled); err != nil {
		return assemblyProducer.ShipAssembled{}, fmt.Errorf("failed to unmarshal shipAssembled: %w", err)
	}

	return assemblyProducer.ShipAssembled{
		EventUuid:    shipAssembled.EventUuid,
		UserUuid:     shipAssembled.UserUuid,
		OrderUuid:    shipAssembled.OrderUuid,
		BuildTimeSec: shipAssembled.BuildTimeSec,
	}, nil
}
