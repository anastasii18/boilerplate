package app

import (
	"context"
	"fmt"
	"inventory/pkg/db"
	rpc "inventory/pkg/grpc"
	"inventory/pkg/service"
	"log"
	"net"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	MongoURI string
	MongoDB  string
	GrpcPort string
}

type App struct {
	Config           *Config
	inventoryApi     *rpc.Api
	inventoryService service.InventoryService
	repository       *db.Repository
	Server           *grpc.Server
}

func New(ctx context.Context, config *Config, database *db.DB, seed bool) *App {
	a := &App{Config: config, Server: grpc.NewServer()}
	a.repository = db.NewRepository(database)

	if seed {
		a.repository.Seed(ctx)
	}

	a.inventoryService = service.NewService(a.repository)

	return a
}

func (a *App) createServer() {
	a.inventoryApi = rpc.New(a.inventoryService)
	inventoryV1.RegisterInventoryServiceServer(a.Server, a.inventoryApi)
}

func (a *App) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.Config.GrpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	a.createServer()
	// Включаем рефлексию для отладки
	reflection.Register(a.Server)

	go func() {
		log.Printf("🚀 gRPC server listening on %s\n", a.Config.GrpcPort)
		err = a.Server.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()
}
