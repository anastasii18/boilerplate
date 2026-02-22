package service

import (
	"fmt"
	"time"
)

type Part struct {
	Uuid          string            // Уникальный идентификатор детали
	Name          string            // Название детали
	Description   string            // Описание детали
	Price         float64           // Цена за единицу
	StockQuantity int64             // Количество на складе
	Category      Category          // Категория
	Dimensions    *Dimensions       // Размеры детали
	Manufacturer  *Manufacturer     // Информация о производителе
	Tags          []string          // Теги для быстрого поиска
	Metadata      map[string]*Value // Гибкие метаданные
	CreatedAt     time.Time         // Дата создания
	UpdatedAt     *time.Time        // Дата обновления
}

type PartSearch struct {
	Uuids                 []string
	Names                 []string
	Categories            []Category
	ManufacturerCountries []string
	Tags                  []string
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
	Length float64 // Длина в см
	Width  float64 // Ширина в см
	Height float64 // Высота в см
	Weight float64 // Вес в кг
}

// Информация о производителе
type Manufacturer struct {
	Name    string // Название
	Country string // Страна производства
	Website string // Сайт производителя
}

// Гибкие метаданные
type Value struct {
	StringVal *string
	IntVal    *int64
	FloatVal  *float64
	BoolVal   *bool
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
