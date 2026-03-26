package app

import (
	"auth/pkg/db"
	"auth/pkg/grpc"
	"auth/pkg/service"
	"context"
	authV1 "shared/pkg/proto/auth/v1"

	grpc2 "google.golang.org/grpc"
)

type diContainer struct {
	authApi     *grpc.Api
	authService service.AuthService
	repository  *db.Repository
	Server      *grpc2.Server
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) NewAuthApi(ctx context.Context, config *Config) *grpc.Api {
	if d.authApi == nil {
		d.authApi = grpc.New(d.NewAuthService(ctx, config))
	}
	return d.authApi
}

func (d *diContainer) NewRepository(ctx context.Context, config *Config) *db.Repository {
	if d.repository == nil {
		database, err := db.NewDB(ctx, config.DbUri)
		if err != nil {
			panic(err)
		}
		d.repository = db.NewRepository(database)
	}

	return d.repository
}

func (d *diContainer) NewAuthService(ctx context.Context, config *Config) service.AuthService {
	if d.authService == nil {
		d.authService = service.NewService(d.NewRepository(ctx, config))
	}

	return d.authService
}

func (d *diContainer) NewServer() *grpc2.Server {
	if d.Server == nil {
		d.Server = grpc2.NewServer()
	}

	return d.Server
}

func (d *diContainer) RegisterGRPCServices() {
	authV1.RegisterAuthServiceServer(d.NewServer(), d.authApi)
}
