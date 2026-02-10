package app

import (
	"fmt"
	"inventory/pkg/db"
	rpc "inventory/pkg/grpc"
	"log"
	"net"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Port int
}

type App struct {
	Config           *Config
	inventoryService *rpc.InventoryService
	Server           *grpc.Server
	dbc              *db.DB
}

func New(port int, dbc *db.DB) *App {
	return &App{Config: &Config{Port: port}, Server: grpc.NewServer(), dbc: dbc}
}

func (a *App) createServer() {
	a.inventoryService = rpc.NewInventoryService(a.dbc)
	inventoryV1.RegisterInventoryServiceServer(a.Server, a.inventoryService)
}

func (a *App) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Config.Port))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	a.createServer()
	// Включаем рефлексию для отладки
	reflection.Register(a.Server)

	go func() {
		log.Printf("🚀 gRPC server listening on %d\n", a.Config.Port)
		err = a.Server.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()
}
