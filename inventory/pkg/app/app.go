package app

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/reflection"
)

type Config struct {
	MongoURI          string
	MongoDB           string
	GrpcPort          string
	ServerAuthAddress string
}

type App struct {
	Config    *Config
	Container *Container
	Cleanup   func()
}

func New(ctx context.Context, config *Config, initIndexes bool, seed bool) *App {
	a := &App{Config: config}
	var err error
	a.Container, a.Cleanup, err = InitializeContainer(ctx, config, initIndexes)
	if err != nil {
		log.Fatal(err)
	}

	if seed {
		a.Container.repository.Seed(ctx)
	}

	return a
}

func (a *App) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.Config.GrpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	a.Container.RegisterGRPCServices()
	// Включаем рефлексию для отладки
	reflection.Register(a.Container.Server)

	go func() {
		log.Printf("🚀 gRPC server listening on %s\n", a.Config.GrpcPort)
		err = a.Container.Server.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()
}
