package service

import (
	"context"
	"inventory/pkg/db"
)

type InventoryService interface {
	GetParts(ctx context.Context, filter PartSearch) (map[string]*Part, error)
	GetPart(ctx context.Context, id string) (*Part, error)
}

var _ InventoryService = (*Service)(nil)

type Service struct {
	InventoryRepository db.InventoryRepository
}

func NewService(repository db.InventoryRepository) *Service {
	return &Service{
		InventoryRepository: repository,
	}
}

func (s Service) GetParts(ctx context.Context, filter PartSearch) (map[string]*Part, error) {
	parts, err := s.InventoryRepository.GetParts(ctx, filter.ToDB())
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Part)
	for id, value := range parts {
		result[id] = NewPart(value)
	}

	return result, nil
}

func (s Service) GetPart(ctx context.Context, id string) (*Part, error) {
	part, err := s.InventoryRepository.GetPart(ctx, id)
	if err != nil {
		return nil, err
	}

	return NewPart(part), nil
}
