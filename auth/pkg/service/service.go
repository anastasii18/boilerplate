package service

import (
	"auth/pkg/db"
	"auth/pkg/db/cache"
	"context"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, user User) (*string, error)
	Login(ctx context.Context, user User) (*string, error)
	GetUser(ctx context.Context, uuid string) (*User, error)
}

type Service struct {
	repo      db.AuthRepository
	cacheRepo cache.UserCacheRepository
	cacheTTL  time.Duration
}

var _ AuthService = (*Service)(nil)

func NewService(authRepository db.AuthRepository, cacheRepo cache.UserCacheRepository, cacheTTL time.Duration) *Service {
	return &Service{
		repo:      authRepository,
		cacheRepo: cacheRepo,
		cacheTTL:  cacheTTL,
	}
}

func (s *Service) Register(ctx context.Context, user User) (*string, error) {
	userUuid := uuid.New().String()
	user.UserUuid = &userUuid

	return s.repo.Register(ctx, UserToRepoModel(user))
}

func (s *Service) Login(ctx context.Context, user User) (*string, error) {
	sessionUuid, err := s.repo.Login(ctx, UserToRepoModel(user))
	if err != nil {
		return nil, err
	}

	userFull, err := s.repo.GetUserByLogin(ctx, user.Login)
	err = s.cacheRepo.Set(ctx, Val(sessionUuid), UserToRedisView(NewUser(userFull)), s.cacheTTL)
	if err != nil {
		return nil, err
	}

	return sessionUuid, nil
}

func (s *Service) GetUser(ctx context.Context, uuid string) (*User, error) {
	serviceUser, err := s.repo.GetUserByUuid(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return NewUser(serviceUser), err
}
