package service

import (
	"inventory/pkg/model"
)

type InventoryService interface {
	GetParts(filter model.Filter) map[string]*model.Part
	GetPart(id string) (*model.Part, error)
}
