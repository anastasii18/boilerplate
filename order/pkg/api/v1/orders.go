package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order/pkg/client/payment"
	"order/pkg/service"
	"platform/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/shopspring/decimal"
)

const urlParam = "order_uuid"

type Api struct {
	orderService service.OrderService
}

func New(service service.OrderService) *Api {
	return &Api{
		orderService: service,
	}
}

// Создаёт новый заказ на основе выбранных пользователем деталей.
func (a *Api) CreateOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Декодируем данные из тела запроса
		var orderCreate service.Order
		if err := json.NewDecoder(r.Body).Decode(&orderCreate); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := a.orderService.CreateOrder(ctx, &orderCreate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		type resultCreate struct {
			TotalPrice decimal.Decimal `json:"total_price"`
			OrderUuid  string          `json:"order_uuid"`
		}

		var result resultCreate
		result.OrderUuid = orderCreate.OrderUuid
		result.TotalPrice = orderCreate.TotalPrice

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, result)
	}
}

// Возвращает информацию о заказе
func (a *Api) GetOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(ctx, orderId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found", orderId), http.StatusNotFound)
			return
		}

		render.JSON(w, r, orderData)
	}
}

// Проводит оплату ранее созданного заказа.
func (a *Api) PayOrderHandler(paymentClient *payment.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(ctx, orderId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found", orderId), http.StatusNotFound)
			return
		}

		// Декодируем данные из тела запроса
		var payBody service.PayBody
		if err := json.NewDecoder(r.Body).Decode(&payBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		payOrderResponse, err := paymentClient.PayOrder(ctx, orderId, orderData.UserUuid, payBody.PaymentMethod)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.orderService.UpdateOrder(ctx, orderId, Ptr(payOrderResponse.GetTransactionUuid()), Ptr(service.PAID), Ptr(payBody.PaymentMethod))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		type PayResponse struct {
			TransactionUuid string `json:"transaction_uuid"`
		}

		render.JSON(w, r, PayResponse{payOrderResponse.GetTransactionUuid()})
	}
}

// Отменяет заказ
func (a *Api) CancelOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required.", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(ctx, orderId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found.", orderId), http.StatusNotFound)
			return
		}

		if orderData.Status == service.COMPLETED {
			http.Error(w, fmt.Sprintf("Order with ID '%s' has already been completed and cannot be cancelled.", orderId), http.StatusConflict)
			return
		}

		if orderData.Status == service.PAID {
			http.Error(w, fmt.Sprintf("Order with ID '%s' has already been paid and cannot be cancelled.", orderId), http.StatusConflict)
			return
		}

		if orderData.Status == service.CANCELLED {
			http.Error(w, fmt.Sprintf("Order with ID '%s' has been cancelled.", orderId), http.StatusConflict)
			return
		}

		err = a.orderService.UpdateOrder(ctx, orderId, nil, Ptr(service.CANCELLED), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		render.NoContent(w, r)
		logger.Info(ctx, fmt.Sprintf("Order with ID '%s' has been successfully cancelled.", orderId))
	}
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
