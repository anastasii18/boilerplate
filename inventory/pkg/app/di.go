//go:build wireinject
// +build wireinject

package app

import (
	"context"
	"inventory/pkg/db"
	rpc "inventory/pkg/grpc"
	"inventory/pkg/service"
	authV1 "shared/pkg/proto/auth/v1"
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

func ProvideGRPCServerOptions(interceptor *rpc.AuthInterceptor) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
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

func ProvideAuthClient(ctx context.Context, cfg *Config) (authV1.AuthServiceClient, func(), error) {
	conn, err := grpc.NewClient(
		cfg.ServerAuthAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := authV1.NewAuthServiceClient(conn)

	cleanup := func() {
		_ = conn.Close()
	}

	return client, cleanup, nil
}

func ProvideAuthInterceptor(authClient authV1.AuthServiceClient) *rpc.AuthInterceptor {
	return rpc.NewAuthInterceptor(authClient)
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
	ProvideAuthClient,
	ProvideAuthInterceptor,
	ProvideGRPCServerOptions,
	ProvideServer,
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

func ProvideServer(opts []grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opts...)
}
