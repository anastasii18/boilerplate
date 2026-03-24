//go:build wireinject
// +build wireinject

package app

import (
	"context"
	"inventory/pkg/db"
	rpc "inventory/pkg/grpc"
	"inventory/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"github.com/google/wire"
	"google.golang.org/grpc"
)

type Container struct {
	inventoryApi     *rpc.Api
	inventoryService service.InventoryService
	repository       *db.Repository
	Server           *grpc.Server
	DB               *db.DB
}

func ProvideGRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		// grpc.UnaryInterceptor(...),
		// grpc.StreamInterceptor(...),
		// grpc.MaxRecvMsgSize(10*1024*1024),
		// grpc.Creds(...),
		// grpc.KeepaliveParams(...),
		// grpc.NumStreamWorkers(...),
	}
}

func ProvideMongoDB(ctx context.Context, cfg *Config) (*db.DB, error) {
	return db.NewDB(ctx, cfg.MongoURI, cfg.MongoDB)
}

func InitializeDBIndexes(ctx context.Context, db *db.DB, initIndexes bool) error {
	return db.InitIndex(ctx, initIndexes)
}

var DatabaseSet = wire.NewSet(
	ProvideMongoDB,
	db.NewRepository,
	wire.Bind(new(db.InventoryRepository), new(*db.Repository)),
	InitializeDBIndexes,
)

var ServiceSet = wire.NewSet(
	service.NewService,
	wire.Bind(new(service.InventoryService), new(*service.Service)),
)

var GRPCServerSet = wire.NewSet(
	ProvideGRPCServerOptions,
	grpc.NewServer,
)

var APISet = wire.NewSet(
	rpc.New,
)

// Главная функция инъекции
func InitializeContainer(ctx context.Context, cfg *Config, initIndexes bool) (*Container, func(), error) {
	panic(wire.Build(
		DatabaseSet,
		ServiceSet,
		GRPCServerSet,
		APISet,
		wire.Struct(new(Container), "*"),
	))
}

func (c *Container) RegisterGRPCServices() {
	inventoryV1.RegisterInventoryServiceServer(c.Server, c.inventoryApi)
}
