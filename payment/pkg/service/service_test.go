package service

import (
	"payment/pkg/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPayOrder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"Проверка генерации",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				paymentRepository: repository.NewRepository(),
			}
			got := service.PayOrder(nil)
			require.NoError(t, uuid.Validate(got))
		})
	}
}
