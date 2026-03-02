package app

import (
	"fmt"
	"inventory/pkg/db"
	rpc "inventory/pkg/grpc"
	"inventory/pkg/service"
	"log"
	"net"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Port int
}

type App struct {
	Config           *Config
	inventoryApi     *rpc.Api
	inventoryService service.InventoryService
	repository       *db.Repository
	Server           *grpc.Server
	MongoClient      *mongo.Client
}

func New(port int) (*App, error) {
	a := &App{Config: &Config{Port: port}, Server: grpc.NewServer()}
	var err error
	a.repository, a.MongoClient, err = db.NewRepoWithMongo()
	a.repository.Seed()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %v\n", err)
	}
	a.inventoryService = service.NewService(a.repository)

	return a, nil
}

func (a *App) createServer() {
	a.inventoryApi = rpc.New(a.inventoryService)
	inventoryV1.RegisterInventoryServiceServer(a.Server, a.inventoryApi)
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
