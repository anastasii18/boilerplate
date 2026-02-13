package repository

import "inventory/pkg/model"

func RepoModelToPart(part *Part) *model.Part {
	var metadata = make(map[string]*model.Value)
	if part.Metadata != nil {
		for key, value := range part.Metadata {
			v := &model.Value{}

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

	var dimensions *model.Dimensions
	if part.Dimensions != nil {
		dimensions = &model.Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturer *model.Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &model.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}
	return &model.Part{
		Uuid:          part.Uuid,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      model.Category(part.Category),
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          part.Tags,
		Metadata:      metadata,
		CreatedAt:     part.CreatedAt,
		UpdatedAt:     part.UpdatedAt,
	}
}
