package cache

import (
	"auth/pkg/db"
	"context"
	"errors"
	"fmt"
	"platform/pkg/cache"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

const cacheKeyPrefix = "auth:user:"

var ErrUserNotFound = errors.New("user not found")

type UserCacheRepository interface {
	Set(ctx context.Context, uuid string, user db.UserRedisView, ttl time.Duration) error
	Get(ctx context.Context, uuid string) (db.UserRedisView, error)
}

type Repository struct {
	cache cache.RedisClient
}

func NewRepository(cache cache.RedisClient) *Repository {
	return &Repository{
		cache: cache,
	}
}

var _ UserCacheRepository = (*Repository)(nil)

func (r *Repository) getCacheKey(uuid string) string {
	return fmt.Sprintf("%s%s", cacheKeyPrefix, uuid)
}

func (r *Repository) Set(ctx context.Context, uuid string, user db.UserRedisView, ttl time.Duration) error {
	cacheKey := r.getCacheKey(uuid)

	err := r.cache.HashSet(ctx, cacheKey, user)
	if err != nil {
		return err
	}

	return r.cache.Expire(ctx, cacheKey, ttl)
}

func (r *Repository) Get(ctx context.Context, uuid string) (db.UserRedisView, error) {
	cacheKey := r.getCacheKey(uuid)

	values, err := r.cache.HGetAll(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redigo.ErrNil) {
			return db.UserRedisView{}, ErrUserNotFound
		}
		return db.UserRedisView{}, err
	}

	if len(values) == 0 {
		return db.UserRedisView{}, ErrUserNotFound
	}

	var userRedisView db.UserRedisView
	err = redigo.ScanStruct(values, &userRedisView)
	if err != nil {
		return db.UserRedisView{}, err
	}

	return userRedisView, nil
}
