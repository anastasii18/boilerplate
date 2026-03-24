package app

import (
	"context"
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

type diContainer struct {
	Repo            *db.Repository
	Server          *http.Server
	OrderService    service.OrderService
	InventoryClient inventory.Client
	PaymentClient   payment.Client
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) NewRepo(ctx context.Context, config *Config) *db.Repository {
	database, err := db.NewDB(ctx, config.DbUri)
	if err != nil {
		panic(err)
	}
	d.Repo = db.NewRepository(database)

	return d.Repo
}

func (d *diContainer) NewServer(ctx context.Context, config *Config) *http.Server {
	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	a := api.New(d.OrderService)
	// Определяем маршруты
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Получить заказ по UUID
		r.Get("/{order_uuid}", a.GetOrderHandler(ctx))
		// Создание заказа
		r.Post("/", a.CreateOrderHandler(ctx))
		// Оплата заказа
		r.Post("/{order_uuid}/pay", a.PayOrderHandler(ctx, d.NewPaymentClient(ctx, config)))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", a.CancelOrderHandler(ctx))
	})

	d.Server = &http.Server{
		Addr:              net.JoinHostPort("localhost", config.HttpPort),
		Handler:           r,
		ReadHeaderTimeout: config.ReadHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}
	return d.Server
}

func (d *diContainer) NewOrderService(ctx context.Context, config *Config) service.OrderService {
	d.OrderService = service.NewService(d.NewRepo(ctx, config), d.NewInventoryClient(ctx, config))
	return d.OrderService
}

func (d *diContainer) NewInventoryClient(ctx context.Context, config *Config) inventory.Client {
	inventoryClient, err := inventory.NewClient(config.ServerInventoryAddress)
	if err != nil {
		panic(err)
	}
	d.InventoryClient = inventoryClient

	return d.InventoryClient
}

func (d *diContainer) NewPaymentClient(ctx context.Context, config *Config) payment.Client {
	paymentClient, err := payment.NewClient(config.ServerPaymentAddress)
	if err != nil {
		panic(err)
	}
	d.PaymentClient = paymentClient

	return d.PaymentClient
}
