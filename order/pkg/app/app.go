package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	api "order/pkg/api/v1"
	"order/pkg/client/inventory"
	"order/pkg/client/payment"
	"order/pkg/db"
	"order/pkg/service"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Config struct {
	Port              string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

type App struct {
	Config          *Config
	Storage         *db.Repository
	Server          *http.Server
	OrderService    service.OrderService
	InventoryClient inventory.Client
	PaymentClient   payment.Client
}

func New(ctx context.Context, config *Config, serverInventoryAddress, serverPaymentAddress string, database *db.DB) *App {
	storage := db.NewRepository(database)
	a := &App{Config: config, Storage: storage, OrderService: service.NewService(storage)}

	a.InventoryClient, _ = inventory.NewClient(serverInventoryAddress)
	a.PaymentClient, _ = payment.NewClient(serverPaymentAddress)
	a.createServer(a.createRouter(ctx))

	return a
}

func (app *App) createRouter(ctx context.Context) *chi.Mux {
	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	a := api.New(app.OrderService)
	// Определяем маршруты
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Получить заказ по UUID
		r.Get("/{order_uuid}", a.GetOrderHandler(ctx))
		// Создание заказа
		r.Post("/", a.CreateOrderHandler(ctx, app.InventoryClient))
		// Оплата заказа
		r.Post("/{order_uuid}/pay", a.PayOrderHandler(ctx, app.PaymentClient))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", a.CancelOrderHandler(ctx))
	})

	return r
}

func (a *App) createServer(r *chi.Mux) {
	a.Server = &http.Server{
		Addr:              net.JoinHostPort("localhost", a.Config.Port),
		Handler:           r,
		ReadHeaderTimeout: a.Config.ReadHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}
}

func (a *App) Start() {

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("🚀 HTTP-сервер запущен на порту %s\n", a.Config.Port)
		err := a.Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("❌ Ошибка запуска сервера: %v\n", err)
		}
	}()
}

func (a *App) Stop() {
	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), a.Config.ShutdownTimeout)
	defer cancel()

	err := a.Server.Shutdown(ctx)
	if err != nil {
		log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("✅ Сервер остановлен")
}
