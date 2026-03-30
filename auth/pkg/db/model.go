package db

type User struct {
	UserUuid            string            `json:"user_uuid"`
	Login               string            `json:"login"`
	Email               string            `json:"email"`
	Password            string            `json:"hashed_password" db:"hashed_password"`
	NotificationMethods map[string]string `json:"notification_methods"`
}

// UserRedisView - модель для хранения в Redis hash map
type UserRedisView struct {
	UserUuid string `redis:"user_uuid"`
	Login    string `redis:"login"`
	Email    string `redis:"email"`
}
