package app

import (
	"auth/pkg/db"
	repoCache "auth/pkg/db/cache"
	"auth/pkg/grpc"
	"auth/pkg/service"
	"context"
	"platform/pkg/cache"
	"platform/pkg/cache/redis"
	"platform/pkg/logger"
	authV1 "shared/pkg/proto/auth/v1"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	grpc2 "google.golang.org/grpc"
)

type diContainer struct {
	authApi         *grpc.Api
	authService     service.AuthService
	repository      *db.Repository
	Server          *grpc2.Server
	authInterceptor *grpc.AuthInterceptor

	cacheRepo   repoCache.UserCacheRepository
	redisPool   *redigo.Pool
	redisClient cache.RedisClient
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
		d.authService = service.NewService(
			d.NewRepository(ctx, config),
			d.NewRepoCache(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisConnectionTimeout, config.RedisAddress),
			config.RedisCacheTTL)
	}

	return d.authService
}

func (d *diContainer) NewServer(config *Config) *grpc2.Server {
	if d.Server == nil {
		interceptor := d.NewAuthInterceptor(config)
		d.Server = grpc2.NewServer(
			grpc2.UnaryInterceptor(interceptor.Unary()),
		)
	}

	return d.Server
}

func (d *diContainer) NewAuthInterceptor(config *Config) *grpc.AuthInterceptor {
	if d.authInterceptor == nil {
		cacheRepo := d.NewRepoCache(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisConnectionTimeout, config.RedisAddress)
		d.authInterceptor = grpc.NewAuthInterceptor(cacheRepo)
	}
	return d.authInterceptor
}

func (d *diContainer) RegisterGRPCServices(config *Config) {
	authV1.RegisterAuthServiceServer(d.NewServer(config), d.authApi)
}

func (d *diContainer) NewRepoCache(maxIdle int, idleTimeout, connectionTimeout time.Duration, address string) repoCache.UserCacheRepository {
	if d.cacheRepo == nil {
		d.cacheRepo = repoCache.NewRepository(d.RedisClient(maxIdle, idleTimeout, connectionTimeout, address))
	}

	return d.cacheRepo
}

func (d *diContainer) RedisPool(maxIdle int, idleTimeout time.Duration, address string) *redigo.Pool {
	if d.redisPool == nil {
		d.redisPool = &redigo.Pool{
			MaxIdle:     maxIdle,
			IdleTimeout: idleTimeout,
			DialContext: func(ctx context.Context) (redigo.Conn, error) {
				return redigo.DialContext(ctx, "tcp", address)
			},
		}
	}

	return d.redisPool
}

func (d *diContainer) RedisClient(maxIdle int, idleTimeout, connectionTimeout time.Duration, address string) cache.RedisClient {
	if d.redisClient == nil {
		d.redisClient = redis.NewClient(d.RedisPool(maxIdle, idleTimeout, address), logger.Logger(), connectionTimeout)
	}

	return d.redisClient
}
