package db

import (
	"maps"
	inventoryV1 "shared/pkg/proto/inventory/v1"
	"slices"
	"sync"
)

type DB struct {
	mu sync.RWMutex
}

func NewDB() *DB {
	return &DB{}
}

var parts = map[string]*inventoryV1.Part{
	"fbb05498-4db6-48c8-b945-3e56f4e5ad04": {
		Uuid:          "fbb05498-4db6-48c8-b945-3e56f4e5ad04",
		Name:          "test name",
		Description:   "test description",
		Price:         112.33,
		StockQuantity: 38,
		Category:      inventoryV1.Category_CATEGORY_FUEL,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
		Tags:          []string{"fuel", "Moscow"},
	},
	"bf802b57-1c7d-41ff-9cb7-ee43dbadbf98": {
		Uuid:          "bf802b57-1c7d-41ff-9cb7-ee43dbadbf98",
		Name:          "two two",
		Description:   "test description",
		Price:         45.45,
		StockQuantity: 7,
		Category:      inventoryV1.Category_CATEGORY_ENGINE,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Rostov", Website: "https://rostov.com"},
		Tags:          []string{"engine", "Rostov"},
	},
	"29a9ab94-c814-4828-9a02-b96598dbe299": {
		Uuid:          "29a9ab94-c814-4828-9a02-b96598dbe299",
		Name:          "three three",
		Description:   "test description",
		Price:         66.77,
		StockQuantity: 90,
		Category:      inventoryV1.Category_CATEGORY_ENGINE,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
		Tags:          []string{"engine", "Moscow"},
	},
}

func (db *DB) GetParts(filter Filter) map[string]*inventoryV1.Part {
	db.mu.RLock()
	defer db.mu.RUnlock()

	newParts := make(map[string]*inventoryV1.Part)
	maps.Copy(newParts, parts)

	if filter.Uuids != nil {
		for key := range newParts {
			if !slices.Contains(filter.Uuids, key) {
				delete(newParts, key)
			}
		}
	}

	if filter.Names != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.Names, value.GetName()) {
				delete(newParts, key)
			}
		}
	}

	if filter.Categories != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.Categories, value.GetCategory()) {
				delete(newParts, key)
			}
		}
	}

	if filter.ManufacturerCountries != nil {
		for key, value := range newParts {
			if !slices.Contains(filter.ManufacturerCountries, value.GetManufacturer().Country) {
				delete(newParts, key)
			}
		}
	}

	if filter.Tags != nil {
		for key, value := range newParts {
			if !isIntersect(filter.Tags, value.GetTags()) {
				delete(newParts, key)
			}
		}
	}

	return newParts
}

func (db *DB) GetPart(uuid string) (*inventoryV1.Part, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	part, exists := parts[uuid]
	return part, exists
}

func isIntersect(slice1, slice2 []string) bool {
	for _, i := range slice1 {
		if !slices.Contains(slice2, i) {
			return false
		}
	}
	return true
}
