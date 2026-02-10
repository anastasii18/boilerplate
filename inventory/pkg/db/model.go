package db

import inventoryV1 "shared/pkg/proto/inventory/v1"

type Filter struct {
	Uuids                 []string
	Names                 []string
	Categories            []inventoryV1.Category
	ManufacturerCountries []string
	Tags                  []string
}

func NewFilter(categories []inventoryV1.Category, uuids, names, manufacturerCountries, tags []string) Filter {
	return Filter{Uuids: uuids, Names: names, Categories: categories, ManufacturerCountries: manufacturerCountries, Tags: tags}
}
