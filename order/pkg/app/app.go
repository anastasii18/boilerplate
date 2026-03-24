package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	logger "platform/pkg"
	"time"
)

type Config struct {
	ReadHeaderTimeout      time.Duration
	ShutdownTimeout        time.Duration
	DbUri                  string
	MigrationsDir          string
	HttpPort               string
	ServerInventoryAddress string
	ServerPaymentAddress   string
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
