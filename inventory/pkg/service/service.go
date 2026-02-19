package service

import (
	"inventory/pkg/db"
)

type InventoryService interface {
	GetParts(filter PartSearch) map[string]*Part
	GetPart(id string) (*Part, error)
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

func (s Service) GetParts(filter PartSearch) map[string]*Part {
	parts := s.InventoryRepository.GetParts(FilterToRepoFilter(filter))
	result := make(map[string]*Part)
	for id, value := range parts {
		result[id] = RepoModelToModel(value)
	}

	return result
}

func (s Service) GetPart(id string) (*Part, error) {
	part, err := s.InventoryRepository.GetPart(id)
	if err != nil {
		return nil, err
	}

	return RepoModelToModel(part), nil
}
