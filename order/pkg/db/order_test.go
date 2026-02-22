package db_test

import (
	"fmt"
	"order/pkg/db"
	repomocks "order/pkg/db/mocks"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestGetOrder(t *testing.T) {
	type args struct {
		OrderID string
	}

	type mockSetupFunc func(t *testing.T, args args, m *repomocks.OrderRepository)

	tests := []struct {
		name       string
		args       args
		orderMock  mockSetupFunc
		want       *db.Order
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Успешно получен существующий заказ",
			args: args{
				OrderID: gofakeit.UUID(),
			},
			orderMock: func(t *testing.T, args args, m *repomocks.OrderRepository) {
				expectedOrder := &db.Order{
					OrderUuid:       args.OrderID,
					Status:          db.PAID,
					TotalPrice:      1499.99,
					PaymentMethod:   db.CARD,
					TransactionUuid: "2aafc0e7-4bc4-4c95-a699-9a6ee4449ddc",
				}

				m.On("GetOrder", args.OrderID).
					Return(expectedOrder, nil).
					Once()
			},
			want: &db.Order{
				Status:          db.PAID,
				TotalPrice:      1499.99,
				PaymentMethod:   db.CARD,
				TransactionUuid: "2aafc0e7-4bc4-4c95-a699-9a6ee4449ddc",
			},
			wantErr: false,
		},
		{
			name: "Запрошен несуществующий заказ",
			args: args{
				OrderID: gofakeit.UUID(),
			},
			orderMock: func(t *testing.T, args args, m *repomocks.OrderRepository) {
				m.On("GetOrder", args.OrderID).
					Return(nil, fmt.Errorf("Не существует такого заказа.")).
					Once()
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "Не существует такого заказа.",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			repoMock := repomocks.NewOrderRepository(t)

			tt.orderMock(t, tt.args, repoMock)

			got, err := repoMock.GetOrder(tt.args.OrderID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				// Проверяем UUID отдельно — он должен совпадать с тем, что запрашивали
				require.Equal(t, tt.args.OrderID, got.OrderUuid)

				// Проверяем остальные поля через want
				require.Equal(t, tt.want.Status, got.Status)
				require.Equal(t, tt.want.TotalPrice, got.TotalPrice)
				require.Equal(t, tt.want.PaymentMethod, got.PaymentMethod)
				require.Equal(t, tt.want.PartUuids, got.PartUuids)
				require.Equal(t, tt.want.TransactionUuid, got.TransactionUuid)
				require.Equal(t, tt.want.UserUuid, got.UserUuid)
			}

			repoMock.AssertExpectations(t)
		})
	}
}

func TestCreateOrder(t *testing.T) {
	type args struct {
		order *db.Order
	}
	type mockSetupFunc func(t *testing.T, args args, m *repomocks.OrderRepository)
	tests := []struct {
		name      string
		args      args
		orderMock mockSetupFunc
	}{
		{
			name: "Успешное создание заказа",
			args: args{&db.Order{OrderUuid: gofakeit.UUID()}},
			orderMock: func(t *testing.T, args args, m *repomocks.OrderRepository) {
				expectedOrder := args.order
				m.On("CreateOrder", args.order).
					Once()
				m.On("GetOrder", args.order.OrderUuid).
					Return(expectedOrder, nil).
					Once()
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			repoMock := repomocks.NewOrderRepository(t)

			tt.orderMock(t, tt.args, repoMock)
			repoMock.CreateOrder(tt.args.order)
			got, err := repoMock.GetOrder(tt.args.order.OrderUuid)

			if got == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repoMock.AssertExpectations(t)
		})
	}
}
