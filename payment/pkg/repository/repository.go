package repository

import (
	"github.com/google/uuid"
)

type PaymentRepository interface {
	PayOrder() string
}

var _ PaymentRepository = (*Repository)(nil)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) PayOrder() string {
	return uuid.New().String()
}
