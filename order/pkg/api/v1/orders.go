package v1

import (
	"encoding/json"
	"fmt"
	inventoryApi "inventory/pkg/grpc"
	inventoryService "inventory/pkg/service"
	"log"
	"net/http"
	"order/pkg/db"
	"order/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"
	paymentV1 "shared/pkg/proto/payment/v1"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/samber/lo"
)

const urlParam = "order_uuid"

type Api struct {
	orderService service.OrderService
}

func New() *Api {
	return &Api{
		orderService: service.NewService(),
	}
}

// Создаёт новый заказ на основе выбранных пользователем деталей.
func (a *Api) CreateOrderHandler(inventoryClient inventoryV1.InventoryServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем данные из тела запроса
		var orderCreate service.Order
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
		var newParts []*inventoryService.Part
		for _, part := range parts.Parts {
			newParts = append(newParts, inventoryApi.PartToModel(part))
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.orderService.CreateOrder(ctx, &orderCreate, newParts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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
func (a *Api) GetOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(orderId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found", orderId), http.StatusNotFound)
			return
		}

		render.JSON(w, r, orderData)
	}
}

// Проводит оплату ранее созданного заказа.
func (a *Api) PayOrderHandler(paymentClient paymentV1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(orderId)
		if err != nil {
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

		err = a.orderService.UpdateOrder(orderId, lo.ToPtr(payOrderResponse.GetTransactionUuid()), lo.ToPtr(service.PAID), lo.ToPtr(service.OrderPaymentMethod(payBody.PaymentMethod)))
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
		orderId := chi.URLParam(r, urlParam)
		if orderId == "" {
			http.Error(w, "OrderID parameter is required.", http.StatusBadRequest)
			return
		}

		orderData, err := a.orderService.GetOrder(orderId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Order with ID '%s' not found.", orderId), http.StatusNotFound)
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

		err = a.orderService.UpdateOrder(orderId, nil, lo.ToPtr(service.CANCELLED), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		render.NoContent(w, r)
		log.Printf("Order with ID '%s' has been successfully cancelled.", orderId)
	}
}
