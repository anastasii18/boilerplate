package repository

import (
	"github.com/google/uuid"
)

var _ PaymentRepository = (*Repository)(nil)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) PayOrder() string {
	return uuid.New().String()
}
