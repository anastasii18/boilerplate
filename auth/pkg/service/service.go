package service

import (
	"auth/pkg/db"
	"context"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, user User) (*string, error)
	Login(ctx context.Context, user User) (*string, error)
	GetUser(ctx context.Context, uuid string) (*User, error)
}

type Service struct {
	repo db.AuthRepository
}

var _ AuthService = (*Service)(nil)

func NewService(authRepository db.AuthRepository) *Service {
	return &Service{
		repo: authRepository,
	}
}

func (s *Service) Register(ctx context.Context, user User) (*string, error) {
	userUuid := uuid.New().String()
	user.UserUuid = &userUuid

	return s.repo.Register(ctx, UserToRepoModel(user))
}

func (s *Service) Login(ctx context.Context, user User) (*string, error) {
	return s.repo.Login(ctx, UserToRepoModel(user))
}

func (s *Service) GetUser(ctx context.Context, uuid string) (*User, error) {
	serviceUser, err := s.repo.GetUserByUuid(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return NewUser(serviceUser), err
}
