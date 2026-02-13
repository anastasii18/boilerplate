package repository

import (
	"fmt"
	"time"
)

type Part struct {
	// Уникальный идентификатор детали
	Uuid string
	// Название детали
	Name string
	// Описание детали
	Description string
	// Цена за единицу
	Price float64
	// Количество на складе
	StockQuantity int64
	// Категория
	Category Category
	// Размеры детали
	Dimensions *Dimensions
	// Информация о производителе
	Manufacturer *Manufacturer
	// Теги для быстрого поиска
	Tags []string
	// Гибкие метаданные
	Metadata map[string]*Value
	// Дата создания
	CreatedAt *time.Time
	// Дата обновления
	UpdatedAt *time.Time
}

type Filter struct {
	Uuids                 []string
	Names                 []string
	Categories            []Category
	ManufacturerCountries []string
	Tags                  []string
}

func NewFilter(categories []Category, uuids, names, manufacturerCountries, tags []string) Filter {
	return Filter{Uuids: uuids, Names: names, Categories: categories, ManufacturerCountries: manufacturerCountries, Tags: tags}
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
	// Длина в см
	Length float64
	// Ширина в см
	Width float64
	// Высота в см
	Height float64
	// Вес в кг
	Weight float64
}

// Информация о производителе
type Manufacturer struct {
	// Название
	Name string
	// Страна производства
	Country string
	// Сайт производителя
	Website string
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
