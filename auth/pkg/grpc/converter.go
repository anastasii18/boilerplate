package grpc

import (
	"auth/pkg/service"
	authV1 "shared/pkg/proto/auth/v1"
)

func NewNotificationMethods(methods []*authV1.NotificationMethod) map[string]string {
	notificationMethods := make(map[string]string)

	for _, method := range methods {
		notificationMethods[method.GetProviderName()] = method.GetTarget()
	}

	return notificationMethods
}

func NotificationMethodsFomMap(methods map[string]string) []*authV1.NotificationMethod {
	var notificationMethods []*authV1.NotificationMethod

	for providerName, target := range methods {
		notificationMethods = append(notificationMethods, &authV1.NotificationMethod{ProviderName: providerName, Target: target})
	}

	return notificationMethods
}

func NewUserByRegister(request *authV1.RegisterRequest) service.User {
	notificationMethods := NewNotificationMethods(request.GetNotificationMethods())

	return service.User{
		Password:            request.GetPassword(),
		Login:               request.GetLogin(),
		Email:               request.GetEmail(),
		NotificationMethods: notificationMethods,
	}
}

func NewUserByLogin(request *authV1.LoginRequest) service.User {
	return service.User{
		Password: request.GetPassword(),
		Login:    request.GetLogin(),
	}
}

func Val[T any, P *T](p P) T {
	if p != nil {
		return *p
	}
	var def T
	return def
}
