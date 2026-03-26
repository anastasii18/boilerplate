package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"platform/pkg/logger"

	"google.golang.org/grpc/reflection"
)

type Config struct {
	GrpcPort      string
	DbUri         string
	MigrationsDir string
}

type App struct {
	Config      *Config
	DiContainer *diContainer
}

func New(ctx context.Context, config *Config) (*App, error) {
	a := &App{Config: config}
	err := a.initDeps(ctx)
	a.DiContainer.NewAuthApi(ctx, config)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (app *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		app.initDI,
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
	app.DiContainer = NewDiContainer()
	return nil
}

// Уровень логирования (debug, info, warn, error)
func (app *App) initLogger(ctx context.Context) error {
	return logger.Init(
		"debug",
		true,
	)
}

func (a *App) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.Config.GrpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	a.DiContainer.RegisterGRPCServices()
	// Включаем рефлексию для отладки
	reflection.Register(a.DiContainer.Server)

	go func() {
		log.Printf("🚀 gRPC server listening on %s\n", a.Config.GrpcPort)
		err = a.DiContainer.Server.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()
}
