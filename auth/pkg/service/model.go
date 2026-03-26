package service

type User struct {
	UserUuid            *string
	Login               string
	Password            string
	Email               string
	NotificationMethods map[string]string // Список каналов уведомлений пользователя
}
