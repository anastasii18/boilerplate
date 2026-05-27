package app

import (
	"context"
	"fmt"
	"platform/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type App struct {
	diContainer *diContainer
	config      *Config
}

type Config struct {
	KafkaBroker      string
	ConsumeTopicName string
	ProduceTopicName string
	ConsumerGroupId  string
}

func New(ctx context.Context, config *Config) (*App, error) {
	a := &App{config: config}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDI,
		a.initLogger,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initDI(ctx context.Context) error {
	a.diContainer = NewDiContainer()
	return nil
}

// Уровень логирования (debug, info, warn, error)
func (app *App) initLogger(ctx context.Context) error {
	return logger.Init(
		"debug",
		true,
	)
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	done := make(chan struct{})

	consumerCtx, consumerCancel := context.WithCancel(ctx)

	go func() {
		defer close(done)
		if err := a.runConsumer(consumerCtx); err != nil {
			errCh <- fmt.Errorf("consumer crashed: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info(ctx, "Shutdown signal received")
	case err := <-errCh:
		logger.Error(ctx, "Component crashed", zap.Error(err))
		consumerCancel()
		<-done
		return err
	}

	// Graceful shutdown
	logger.Info(ctx, "Initiating graceful shutdown...")
	consumerCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 15*time.Second)
	defer shutdownCancel()

	select {
	case <-done:
		logger.Info(shutdownCtx, "Consumer stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn(shutdownCtx, "Graceful shutdown timeout")
	}

	return nil
}

func (a *App) runConsumer(ctx context.Context) error {
	logger.Info(ctx, "🚀 Assembly Kafka consumer running")

	consumerService, err := a.diContainer.ConsumerService(a.config.ProduceTopicName, a.config.ConsumeTopicName, a.config.KafkaBroker, a.config.ConsumerGroupId)
	if err != nil {
		return err
	}

	err = consumerService.RunConsumer(ctx)
	if err != nil {
		return err
	}

	return nil
}
