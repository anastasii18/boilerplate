package service_test

import (
	"fmt"
	"order/pkg/service"
	"order/pkg/service/mocks"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestGetOrder(t *testing.T) {
	type args struct {
		orderUuid string
	}
	type mockSetupFunc func(t *testing.T, args args, m *mocks.OrderService)
	tests := []struct {
		name      string
		args      args
		want      *service.Order
		orderMock mockSetupFunc
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Успешно получен существующий заказ",
			args: args{
				orderUuid: gofakeit.UUID(),
			},
			orderMock: func(t *testing.T, args args, m *mocks.OrderService) {
				expectedOrder := &service.Order{
					OrderUuid:       args.orderUuid,
					Status:          service.PAID,
					TotalPrice:      1499.99,
					PaymentMethod:   service.CARD,
					TransactionUuid: Ptr("2aafc0e7-4bc4-4c95-a699-9a6ee4449ddc"),
				}

				m.On("GetOrder", args.orderUuid).
					Return(expectedOrder, nil).
					Once()
			},
			want: &service.Order{
				Status:          service.PAID,
				TotalPrice:      1499.99,
				PaymentMethod:   service.CARD,
				TransactionUuid: Ptr("2aafc0e7-4bc4-4c95-a699-9a6ee4449ddc"),
			},
			wantErr: false,
		},
		{
			name: "Запрошен несуществующий заказ",
			args: args{
				orderUuid: gofakeit.UUID(),
			},
			orderMock: func(t *testing.T, args args, m *mocks.OrderService) {
				m.On("GetOrder", args.orderUuid).
					Return(nil, fmt.Errorf("Не существует такого заказа.")).
					Once()
			},
			want:    nil,
			wantErr: true,
			errMsg:  "Не существует такого заказа.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewOrderService(t)
			tt.orderMock(t, tt.args, serviceMock)

			got, err := serviceMock.GetOrder(tt.args.orderUuid)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				// Проверяем UUID отдельно — он должен совпадать с тем, что запрашивали
				require.Equal(t, tt.args.orderUuid, got.OrderUuid)

				// Проверяем остальные поля через want
				require.Equal(t, tt.want.Status, got.Status)
				require.Equal(t, tt.want.TotalPrice, got.TotalPrice)
				require.Equal(t, tt.want.PaymentMethod, got.PaymentMethod)
				require.Equal(t, tt.want.PartUuids, got.PartUuids)
				require.Equal(t, tt.want.TransactionUuid, got.TransactionUuid)
				require.Equal(t, tt.want.UserUuid, got.UserUuid)
			}
		})
	}
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
