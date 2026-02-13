package repository

import (
	"inventory/pkg/model"
)

type InventoryRepository interface {
	GetParts(filter Filter) map[string]*model.Part
	GetPart(id string) (*model.Part, error)
}
