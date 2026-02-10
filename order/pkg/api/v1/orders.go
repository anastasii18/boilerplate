package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order/pkg/db"
	inventoryV1 "shared/pkg/proto/inventory/v1"
	paymentV1 "shared/pkg/proto/payment/v1"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const urlParam = "order_uuid"

// Создаёт новый заказ на основе выбранных пользователем деталей.
func CreateOrderHandler(storage *db.OrderStorage, inventoryClient inventoryV1.InventoryServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем данные из тела запроса
		var orderCreate db.Order
		if err := json.NewDecoder(r.Body).Decode(&orderCreate); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		getListPartsRequest := &inventoryV1.GetListPartsRequest{
			Filter: &inventoryV1.PartsFilter{
				Uuids: orderCreate.PartUuids,
			},
		}
		ctx := r.Context()
		parts, err := inventoryClient.GetListParts(ctx, getListPartsRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Проверяет, что все детали существуют. Если хотя бы одной нет — возвращает ошибку
		if len(orderCreate.PartUuids) != len(parts.Parts) {
			http.Error(w, "One of part not found", http.StatusNotFound)
			return
		}
		totalPrice := 0.0
		for _, part := range parts.Parts {
			if part.StockQuantity < 1 {
				http.Error(w, "One of part stock quantity not found", http.StatusNotFound)
			}
			totalPrice += part.Price
		}

		orderCreate.TotalPrice = totalPrice
		orderCreate.OrderUuid = uuid.New().String()
		orderCreate.Status = db.PENDING_PAYMENT

		storage.CreateOrder(&orderCreate)

		type resultCreate struct {
			TotalPrice float64 `json:"total_price"`
			OrderUuid  string  `json:"order_uuid"`
		}

		var result resultCreate
		result.OrderUuid = orderCreate.OrderUuid
		result.TotalPrice = orderCreate.TotalPrice

		render.JSON(w, r, result)
	}
}

// Возвращает информацию о заказе
func GetOrderHandler(storage *db.OrderStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData := storage.GetOrder(orderId)
		if orderData == nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found", orderId), http.StatusNotFound)
			return
		}

		render.JSON(w, r, orderData)
	}
}

// Проводит оплату ранее созданного заказа.
func PayOrderHandler(storage *db.OrderStorage, paymentClient paymentV1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData := storage.GetOrder(orderId)
		if orderData == nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found", orderId), http.StatusNotFound)
			return
		}

		type PayBody struct {
			PaymentMethod db.OrderPaymentMethod `json:"payment_method"`
		}

		// Декодируем данные из тела запроса
		var payBody PayBody
		if err := json.NewDecoder(r.Body).Decode(&payBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		payOrderRequest := &paymentV1.PayOrderRequest{
			OrderUuid:     orderId,
			UserUuid:      orderData.UserUuid,
			PaymentMethod: paymentV1.PaymentMethod(payBody.PaymentMethod),
		}
		ctx := r.Context()
		payOrderResponse, err := paymentClient.PayOrder(ctx, payOrderRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		orderData.TransactionUuid = payOrderResponse.GetTransactionUuid()
		orderData.Status = db.PAID
		orderData.PaymentMethod = payBody.PaymentMethod

		type PayResponse struct {
			TransactionUuid string `json:"transaction_uuid"`
		}

		render.JSON(w, r, PayResponse{payOrderResponse.GetTransactionUuid()})
	}
}

// Отменяет заказ
func CancelOrderHandler(storage *db.OrderStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required.", http.StatusBadRequest)
			return
		}

		orderData := storage.GetOrder(orderId)
		if orderData == nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found.", orderId), http.StatusNotFound)
			return
		}

		if orderData.Status == db.PAID {
			http.Error(w, fmt.Sprintf("Order with ID '%s' has already been paid and cannot be cancelled.", orderId), http.StatusConflict)
			return
		}

		if orderData.Status == db.CANCELLED {
			http.Error(w, fmt.Sprintf("Order with ID '%s' has been cancelled.", orderId), http.StatusConflict)
			return
		}

		orderData.Status = db.CANCELLED
		render.NoContent(w, r)
		log.Printf("Order with ID '%s' has been successfully cancelled.", orderId)
	}
}
