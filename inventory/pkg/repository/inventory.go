package repository

import (
	"fmt"
	"inventory/pkg/model"
	"maps"
	"slices"
	"sync"
)

type Repository struct {
	mu   sync.RWMutex
	data map[string]*Part
}

var _ InventoryRepository = (*Repository)(nil)

func (r *Repository) GetParts(filter Filter) map[string]*model.Part {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newParts := make(map[string]*Part)
	maps.Copy(newParts, r.data)

	if filter.Uuids != nil {
		for key := range newParts {
			if !slices.Contains(filter.Uuids, key) {
				delete(newParts, key)
			}
		}
	}

	if filter.Names != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.Names, value.Name) {
				delete(newParts, key)
			}
		}
	}

	if filter.Categories != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.Categories, value.Category) {
				delete(newParts, key)
			}
		}
	}

	if filter.ManufacturerCountries != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.ManufacturerCountries, value.Manufacturer.Country) {
				delete(newParts, key)
			}
		}
	}

	if filter.Tags != nil {
		for key, value := range newParts {
			if !isIntersect(filter.Tags, value.Tags) {
				delete(newParts, key)
			}
		}
	}

	result := make(map[string]*model.Part)
	for id, value := range newParts {
		result[id] = RepoModelToPart(value)
	}

	return result
}

func isIntersect(slice1, slice2 []string) bool {
	for _, i := range slice1 {
		if !slices.Contains(slice2, i) {
			return false
		}
	}
	return true
}

func (r *Repository) GetPart(id string) (*model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	part, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("part with id %s not found", id)
	}

	return RepoModelToPart(part), nil
}

func NewRepository() *Repository {
	return &Repository{
		data: map[string]*Part{
			"fbb05498-4db6-48c8-b945-3e56f4e5ad04": {
				Uuid:          "fbb05498-4db6-48c8-b945-3e56f4e5ad04",
				Name:          "test name",
				Description:   "test description",
				Price:         112.33,
				StockQuantity: 38,
				Category:      CATEGORY_FUEL,
				Manufacturer:  &Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
				Tags:          []string{"fuel", "Moscow"},
			},
			"bf802b57-1c7d-41ff-9cb7-ee43dbadbf98": {
				Uuid:          "bf802b57-1c7d-41ff-9cb7-ee43dbadbf98",
				Name:          "two two",
				Description:   "test description",
				Price:         45.45,
				StockQuantity: 7,
				Category:      CATEGORY_ENGINE,
				Manufacturer:  &Manufacturer{Name: "test name", Country: "Rostov", Website: "https://rostov.com"},
				Tags:          []string{"engine", "Rostov"},
			},
			"29a9ab94-c814-4828-9a02-b96598dbe299": {
				Uuid:          "29a9ab94-c814-4828-9a02-b96598dbe299",
				Name:          "three three",
				Description:   "test description",
				Price:         66.77,
				StockQuantity: 90,
				Category:      CATEGORY_ENGINE,
				Manufacturer:  &Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
				Tags:          []string{"engine", "Moscow"},
			},
		},
	}
}
