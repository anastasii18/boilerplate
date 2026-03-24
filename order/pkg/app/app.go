package app

import (
	"context"
	"fmt"
	"net/http"
	"platform/pkg/logger"
	"time"

	"github.com/go-faster/errors"

	"go.uber.org/zap"
)

type Config struct {
	ReadHeaderTimeout      time.Duration
	ShutdownTimeout        time.Duration
	DbUri                  string
	MigrationsDir          string
	HttpPort               string
	ServerInventoryAddress string
	ServerPaymentAddress   string
	KafkaBroker            string
	ProduceTopicName       string
	ConsumeTopicName       string
	ConsumerGroupId        string
}

type App struct {
	Config      *Config
	diContainer *diContainer
}

func New(ctx context.Context, config *Config) (*App, error) {
	a := &App{Config: config}
	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (app *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		app.initDI,
		app.initServer,
		app.initLogger,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) initDI(ctx context.Context) error {
	app.diContainer = NewDiContainer()
	return nil
}

// Уровень логирования (debug, info, warn, error)
func (app *App) initLogger(ctx context.Context) error {
	return logger.Init(
		"debug",
		true,
	)
}

func (app *App) initServer(ctx context.Context) error {
	app.diContainer.NewOrderService(ctx, app.Config)
	app.diContainer.NewServer(ctx, app.Config)
	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info(ctx, fmt.Sprintf("🚀 HTTP-сервер запущен на порту %s\n", app.Config.HttpPort))
		err := app.diContainer.Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Info(ctx, fmt.Sprintf("❌ Ошибка запуска сервера: %v\n", err))
		}
	}()

	return nil
}

func (app *App) Run(ctx context.Context, config *Config) error {
	// Канал для ошибок от компонентов
	errCh := make(chan error, 2)

	// Контекст для остановки всех горутин
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Консьюмер
	go func() {
		if err := app.runConsumer(ctx, config); err != nil {
			errCh <- errors.Errorf("consumer crashed: %v", err)
		}
	}()

	// Ожидание либо ошибки, либо завершения контекста (например, сигнал SIGINT/SIGTERM)
	select {
	case <-ctx.Done():
		logger.Info(ctx, "Shutdown signal received")
	case err := <-errCh:
		logger.Error(ctx, "Component crashed, shutting down", zap.Error(err))
		// Триггерим cancel, чтобы остановить второй компонент
		cancel()
		// Дождись завершения всех задач (если есть graceful shutdown внутри)
		<-ctx.Done()
		return err
	}

	return nil
}

func (app *App) runConsumer(ctx context.Context, config *Config) error {
	logger.Info(ctx, "🚀 Order Kafka consumer running")

	err := app.diContainer.ConsumerService(config.ConsumeTopicName, config.KafkaBroker, config.ConsumerGroupId).RunConsumer(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) Stop() {
	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), app.Config.ShutdownTimeout)
	defer cancel()

	err := app.diContainer.Server.Shutdown(ctx)
	if err != nil {
		logger.Info(ctx, fmt.Sprintf("❌ Ошибка при остановке сервера: %v\n", err))
	}

	logger.Info(ctx, "✅ Сервер остановлен")
}
