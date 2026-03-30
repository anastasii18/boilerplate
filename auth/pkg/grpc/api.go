package grpc

import (
	"auth/pkg/db"
	"auth/pkg/service"
	"context"
	authV1 "shared/pkg/proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Api struct {
	authV1.UnimplementedAuthServiceServer
	authService service.AuthService
}

func New(authService service.AuthService) *Api {
	return &Api{
		authService: authService,
	}
}

func (a *Api) Register(ctx context.Context, request *authV1.RegisterRequest) (*authV1.RegisterResponse, error) {
	userUuid, err := a.authService.Register(ctx, NewUserByRegister(request))

	if err != nil {
		return nil, err
	}

	return &authV1.RegisterResponse{
		UserUuid: service.Val(userUuid),
	}, nil
}

func (a *Api) Login(ctx context.Context, request *authV1.LoginRequest) (*authV1.LoginResponse, error) {
	sessionUuid, err := a.authService.Login(ctx, NewUserByLogin(request))

	if err != nil {
		return nil, err
	}

	return &authV1.LoginResponse{
		SessionUuid: service.Val(sessionUuid),
	}, nil
}

func (a *Api) GetUser(ctx context.Context, request *authV1.GetUserRequest) (*authV1.GetUserResponse, error) {
	user, err := a.authService.GetUser(ctx, request.GetUserUuid())

	if err != nil {
		return nil, err
	}

	return &authV1.GetUserResponse{
		UserUuid:            Val(user.UserUuid),
		Login:               user.Login,
		Email:               user.Email,
		NotificationMethods: NotificationMethodsFomMap(user.NotificationMethods),
	}, nil
}

func (a *Api) Whoami(ctx context.Context, _ *authV1.WhoamiRequest) (*authV1.WhoamiResponse, error) {
	userView, ok := ctx.Value(UserContextKey).(db.UserRedisView)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get user from context")
	}

	return &authV1.WhoamiResponse{
		UserUuid: userView.UserUuid,
		Login:    userView.Login,
		Email:    userView.Email,
	}, nil
}
