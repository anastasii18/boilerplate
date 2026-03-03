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
	"order/pkg/migrator"
	"order/pkg/service"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
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

func New(config *Config, serverInventoryAddress, serverPaymentAddress string, pool *pgxpool.Pool) *App {
	storage := db.NewRepository(pool)
	a := &App{Config: config, Storage: storage, OrderService: service.NewService(storage)}

	a.InventoryClient, _ = inventory.NewClient(serverInventoryAddress)
	a.PaymentClient, _ = payment.NewClient(serverPaymentAddress)
	a.createServer(a.createRouter())

	return a
}

func (app *App) createRouter() *chi.Mux {
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
		r.Get("/{order_uuid}", a.GetOrderHandler())
		// Создание заказа
		r.Post("/", a.CreateOrderHandler(app.InventoryClient))
		// Оплата заказа
		r.Post("/{order_uuid}/pay", a.PayOrderHandler(app.PaymentClient))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", a.CancelOrderHandler())
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

func (a *App) Migrate(ctx context.Context) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed to load .env file: %v\n", err)
		return
	}

	dbURI := os.Getenv("DB_URI")

	// Создаем соединение с базой данных
	con, err := pgx.Connect(ctx, dbURI)
	if err != nil {
		log.Printf("failed to connect to database: %v\n", err)
		return
	}
	defer func() {
		cerr := con.Close(ctx)
		if cerr != nil {
			log.Printf("failed to close connection: %v\n", cerr)
		}
	}()

	// Проверяем, что соединение с базой установлено
	err = con.Ping(ctx)
	if err != nil {
		log.Printf("База данных недоступна: %v\n", err)
		return
	}

	// Инициализируем мигратор
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	migratorRunner := migrator.NewMigrator(stdlib.OpenDB(*con.Config().Copy()), migrationsDir)

	err = migratorRunner.Up()
	if err != nil {
		log.Printf("Ошибка миграции базы данных: %v\n", err)
		return
	}
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
