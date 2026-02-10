package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order/pkg/db"

	inventoryV1 "shared/pkg/proto/inventory/v1"
	paymentV1 "shared/pkg/proto/payment/v1"
)

const (
	httpPort               = "8080"
	urlParamOrder          = "order_uuid"
	serverInventoryAddress = "localhost:50051"
	serverPaymentAddress   = "localhost:50052"
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// Создаем хранилище для данных
	storage := db.NewOrderStorage()

	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// inventory client
	conn, err := grpc.NewClient(
		serverInventoryAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("failed to close connect: %v", cerr)
		}
	}()

	// Создаем gRPC inventory клиент
	inventoryClient := inventoryV1.NewInventoryServiceClient(conn)

	// payment client
	connPayment, errPayment := grpc.NewClient(
		serverPaymentAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if errPayment != nil {
		log.Printf("failed to connect: %v\n", errPayment)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("failed to close connect: %v", cerr)
		}
	}()

	// Создаем gRPC payment клиент
	paymentClient := paymentV1.NewPaymentServiceClient(connPayment)

	// Определяем маршруты
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Получить заказ по UUID
		r.Get("/{order_uuid}", getOrderHandler(storage))
		// Создание заказа
		r.Post("/", createOrderHandler(storage, inventoryClient))
		// Оплата заказа
		r.Post("/{order_uuid}/pay", payOrderHandler(storage, paymentClient))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", cancelOrderHandler(storage))
	})

	// Запускаем HTTP-сервер
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("🚀 HTTP-сервер запущен на порту %s\n", httpPort)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("❌ Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("✅ Сервер остановлен")
}

// Создаёт новый заказ на основе выбранных пользователем деталей.
func createOrderHandler(storage *db.OrderStorage, inventoryClient inventoryV1.InventoryServiceClient) http.HandlerFunc {
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
func getOrderHandler(storage *db.OrderStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParamOrder)
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
func payOrderHandler(storage *db.OrderStorage, paymentClient paymentV1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParamOrder)
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
func cancelOrderHandler(storage *db.OrderStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderId := chi.URLParam(r, urlParamOrder)
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
