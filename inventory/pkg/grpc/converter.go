package grpc

import (
	"inventory/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewPart(part *service.Part) *inventoryV1.Part {
	category := inventoryV1.Category(part.Category)

	var dimensions *inventoryV1.Dimensions
	if part.Dimensions != nil {
		dimensions = &inventoryV1.Dimensions{
			Length: wrapperspb.Double(part.Dimensions.Length),
			Width:  wrapperspb.Double(part.Dimensions.Width),
			Height: wrapperspb.Double(part.Dimensions.Height),
			Weight: wrapperspb.Double(part.Dimensions.Weight),
		}
	}

	var manufacturer *inventoryV1.Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &inventoryV1.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	var metadata = make(map[string]*inventoryV1.Value)
	if part.Metadata != nil {
		for _, value := range part.Metadata {
			var v inventoryV1.Value
			switch {
			case value.StringVal != nil:
				v.MetaValue = &inventoryV1.Value_StringValue{
					StringValue: wrapperspb.String(*value.StringVal),
				}
			case value.IntVal != nil:
				v.MetaValue = &inventoryV1.Value_Int64Value{
					Int64Value: wrapperspb.Int64(*value.IntVal),
				}
			case value.FloatVal != nil:
				v.MetaValue = &inventoryV1.Value_DoubleValue{
					DoubleValue: wrapperspb.Double(*value.FloatVal),
				}
			case value.BoolVal != nil:
				v.MetaValue = &inventoryV1.Value_BoolValue{
					BoolValue: wrapperspb.Bool(*value.BoolVal),
				}

			default:
				// если все поля nil — возвращаем пустой Value (или можно nil, по вкусу)
			}
		}
	}

	var createdAt *timestamppb.Timestamp
	if !part.CreatedAt.IsZero() {
		createdAt = timestamppb.New(part.CreatedAt)
	}

	var updatedAt *timestamppb.Timestamp
	if part.UpdatedAt != nil && !part.UpdatedAt.IsZero() {
		updatedAt = timestamppb.New(*part.UpdatedAt)
	}

	return &inventoryV1.Part{
		Uuid:          part.Uuid,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      category,
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          part.Tags,
		Metadata:      metadata,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func NewServicePart(part *inventoryV1.Part) *service.Part {
	category := service.Category(part.Category.Number())

	var dimensions *service.Dimensions
	if part.Dimensions != nil {
		dimensions = &service.Dimensions{
			Length: part.Dimensions.GetLength().Value,
			Width:  part.Dimensions.GetWidth().Value,
			Height: part.Dimensions.GetHeight().Value,
			Weight: part.Dimensions.GetWeight().Value,
		}
	}

	var manufacturer *service.Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &service.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	var metadata = make(map[string]*service.Value)
	if part.Metadata != nil {
		for key, protoValue := range part.Metadata {
			var v *service.Value

			// Проверяем, какой именно oneof установлен в protobuf Value
			if protoValue != nil {
				switch x := protoValue.GetMetaValue().(type) {
				case *inventoryV1.Value_StringValue:
					v.StringVal = &x.StringValue.Value
				case *inventoryV1.Value_Int64Value:
					v.IntVal = &x.Int64Value.Value
				case *inventoryV1.Value_DoubleValue:
					v.FloatVal = &x.DoubleValue.Value
				case *inventoryV1.Value_BoolValue:
					v.BoolVal = &x.BoolValue.Value
				default:
					// Если oneof не установлен (nil или неизвестный тип) — оставляем v пустым
				}
			}

			metadata[key] = v
		}
	}

	var createdAt time.Time
	if part.CreatedAt != nil {
		createdAt = part.CreatedAt.AsTime()
	}

	var updatedAt *time.Time
	if part.UpdatedAt != nil {
		updatedAt = Ptr(part.UpdatedAt.AsTime())
	}

	return &service.Part{
		Uuid:          part.Uuid,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      category,
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          part.Tags,
		Metadata:      metadata,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func NewPartSearch(filter *inventoryV1.PartsFilter) service.PartSearch {
	var categories []service.Category
	for _, category := range filter.Categories {
		categories = append(categories, service.Category(category))
	}

	return service.NewPartSearch(categories, filter.Uuids, filter.Names, filter.ManufacturerCountries, filter.Tags)
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
