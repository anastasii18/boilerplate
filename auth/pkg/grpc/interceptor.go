package grpc

import (
	"auth/pkg/db/cache"
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключ для хранения данных пользователя в контексте
type contextKey string

const UserContextKey contextKey = "user_info"

type AuthInterceptor struct {
	cacheRepo cache.UserCacheRepository
}

func NewAuthInterceptor(cacheRepo cache.UserCacheRepository) *AuthInterceptor {
	return &AuthInterceptor{
		cacheRepo: cacheRepo,
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
		// 1. Белый список методов (не требующих авторизации)
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 2. Извлечение метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		// 3. Получение session-uuid
		values := md.Get("session-uuid")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "session-uuid is required")
		}
		sessionUUID := values[0]

		// 4. Проверка сессии в Redis
		userView, err := i.cacheRepo.Get(ctx, sessionUUID)
		if err != nil {
			// Если сессия не найдена или истек TTL
			return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
		}

		// 5. Обогащение контекста данными пользователя для бизнес-логики
		newCtx := context.WithValue(ctx, UserContextKey, userView)

		return handler(newCtx, req)
	}
}

// isPublicMethod проверяет, нужно ли пропускать проверку для данного метода
func isPublicMethod(fullMethod string) bool {
	publicMethods := []string{
		"/auth.v1.AuthService/Login",
		"/auth.v1.AuthService/Register",
	}

	for _, m := range publicMethods {
		if strings.HasSuffix(fullMethod, m) {
			return true
		}
	}
	return false
}
