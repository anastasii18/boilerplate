package service

import (
	"inventory/pkg/converter"
	"inventory/pkg/model"
	"inventory/pkg/repository"
)

var _ InventoryService = (*Service)(nil)

type Service struct {
	InventoryRepository repository.InventoryRepository
}

func (s Service) GetParts(filter model.Filter) map[string]*model.Part {
	return s.InventoryRepository.GetParts(converter.FilterToRepoFilter(filter))
}

func (s Service) GetPart(id string) (*model.Part, error) {
	return s.InventoryRepository.GetPart(id)
}

func NewService() *Service {
	return &Service{
		InventoryRepository: repository.NewRepository(),
	}
}
