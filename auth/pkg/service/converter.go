package service

import "auth/pkg/db"

func NewUser(user *db.User) *User {
	return &User{
		UserUuid:            Ptr(user.UserUuid),
		Email:               user.Email,
		Password:            user.Password,
		Login:               user.Login,
		NotificationMethods: user.NotificationMethods,
	}
}

func UserToRepoModel(user User) *db.User {
	return &db.User{
		UserUuid:            Val(user.UserUuid),
		Email:               user.Email,
		Password:            user.Password,
		Login:               user.Login,
		NotificationMethods: user.NotificationMethods,
	}
}

func Val[T any, P *T](p P) T {
	if p != nil {
		return *p
	}
	var def T
	return def
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
