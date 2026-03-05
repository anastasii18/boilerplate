package db

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Part struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Uuid          string             `bson:"uuid"`           // Уникальный идентификатор детали
	Name          string             `bson:"name"`           // Название детали
	Description   string             `bson:"description"`    // Описание детали
	Price         float64            `bson:"price"`          // Цена за единицу
	StockQuantity int64              `bson:"stock_quantity"` // Количество на складе
	Category      Category           `bson:"category"`       // Категория
	Dimensions    *Dimensions        `bson:"dimensions"`     // Размеры детали
	Manufacturer  *Manufacturer      `bson:"manufacturer"`   // Информация о производителе
	Tags          []string           `bson:"tags"`           // Теги для быстрого поиска
	Metadata      map[string]*Value  `bson:"metadata"`       // Гибкие метаданные
	CreatedAt     time.Time          `bson:"created_at"`     // Дата создания
	UpdatedAt     *time.Time         `bson:"updated_at"`     // Дата обновления
}

type PartSearch struct {
	Uuids                 []string   `json:"uuids,omitempty"`
	Names                 []string   `json:"names,omitempty"`
	Categories            []Category `json:"categories,omitempty"`
	ManufacturerCountries []string   `json:"manufacturer_countries,omitempty"`
	Tags                  []string   `json:"tags,omitempty"`
}

func NewPartSearch(categories []Category, uuids, names, manufacturerCountries, tags []string) PartSearch {
	return PartSearch{Uuids: uuids, Names: names, Categories: categories, ManufacturerCountries: manufacturerCountries, Tags: tags}
}

// Категория
type Category int32

const (
	CATEGORY_UNSPECIFIED Category = 0
	CATEGORY_ENGINE      Category = 1
	CATEGORY_FUEL        Category = 2
	CATEGORY_PORTHOLE    Category = 3
	CATEGORY_WING        Category = 4
)

// Размеры детали
type Dimensions struct {
	Length float64 `bson:"length"` // Длина в см
	Width  float64 `bson:"width"`  // Ширина в см
	Height float64 `bson:"height"` // Высота в см
	Weight float64 `bson:"weight"` // Вес в кг
}

// Информация о производителе
type Manufacturer struct {
	Name    string `bson:"name"`    // Название
	Country string `bson:"country"` // Страна производства
	Website string `bson:"website"` // Сайт производителя
}

// Гибкие метаданные
type Value struct {
	StringVal *string  `bson:"string_val"`
	IntVal    *int64   `bson:"int_val"`
	FloatVal  *float64 `bson:"float_val"`
	BoolVal   *bool    `bson:"bool_val"`
}

func (v Value) IsSet() bool {
	return v.StringVal != nil || v.IntVal != nil || v.FloatVal != nil || v.BoolVal != nil
}

func (v Value) String() string {
	switch {
	case v.StringVal != nil:
		return fmt.Sprintf("string(%q)", *v.StringVal)
	case v.IntVal != nil:
		return fmt.Sprintf("int64(%d)", *v.IntVal)
	case v.FloatVal != nil:
		return fmt.Sprintf("float64(%f)", *v.FloatVal)
	case v.BoolVal != nil:
		return fmt.Sprintf("bool(%t)", *v.BoolVal)
	default:
		return "unset"
	}
}
