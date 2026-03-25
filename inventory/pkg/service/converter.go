package service

import (
	"inventory/pkg/db"
)

func (ps *PartSearch) ToDB() db.PartSearch {
	var categories []db.Category
	for _, category := range ps.Categories {
		categories = append(categories, db.Category(category))
	}

	return db.NewPartSearch(categories, ps.Uuids, ps.Names, ps.ManufacturerCountries, ps.Tags)
}

func NewPart(part *db.Part) *Part {
	var metadata = make(map[string]*Value)
	if part.Metadata != nil {
		for key, value := range part.Metadata {
			v := &Value{}

			if value.IntVal != nil {
				v.IntVal = value.IntVal
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
