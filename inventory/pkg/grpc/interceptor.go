package grpc

import (
	"context"
	authV1 "shared/pkg/proto/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	authClient authV1.AuthServiceClient
}

func NewAuthInterceptor(authClient authV1.AuthServiceClient) *AuthInterceptor {
	return &AuthInterceptor{
		authClient: authClient,
	}
}

// Unary возвращает функцию интерцептора для одиночных запросов
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		//  Извлечение метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		// 3. Получение session-uuid
		values := md.Get("session-uuid")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "session-uuid is required")
		}

		// 4. Проверка сессии в Redis
		outgoingCtx := metadata.NewOutgoingContext(ctx, md)
		_, err := i.authClient.Whoami(outgoingCtx, &authV1.WhoamiRequest{})

		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
		}

		return handler(ctx, req)
	}
}
