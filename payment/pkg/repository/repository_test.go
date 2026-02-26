package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPayOrder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"Тест генерации",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repository{}
			got := r.PayOrder()
			require.NoError(t, uuid.Validate(got))
		})
	}
}
