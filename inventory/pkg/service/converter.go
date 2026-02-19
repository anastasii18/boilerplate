package service

import (
	repomodel "inventory/pkg/db"
)

func FilterToRepoFilter(filter PartSearch) repomodel.PartSearch {
	var categories []repomodel.Category
	for _, category := range filter.Categories {
		categories = append(categories, repomodel.Category(category))
	}

	return repomodel.NewPartSearch(categories, filter.Uuids, filter.Names, filter.ManufacturerCountries, filter.Tags)
}

func RepoModelToModel(part *repomodel.Part) *Part {
	var metadata = make(map[string]*Value)
	if part.Metadata != nil {
		for key, value := range part.Metadata {
			v := &Value{}

			if value.StringVal != nil {
				v.StringVal = value.StringVal
			} else if value.IntVal != nil {
				v.IntVal = value.IntVal
			} else if v.FloatVal != nil {
				v.FloatVal = v.FloatVal
			} else if value.BoolVal != nil {
				v.BoolVal = value.BoolVal
			}

			metadata[key] = v
		}
	}

	var dimensions *Dimensions
	if part.Dimensions != nil {
		dimensions = &Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturer *Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}
	return &Part{
		Uuid:          part.Uuid,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      Category(part.Category),
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          part.Tags,
		Metadata:      metadata,
		CreatedAt:     part.CreatedAt,
		UpdatedAt:     part.UpdatedAt,
	}
}
