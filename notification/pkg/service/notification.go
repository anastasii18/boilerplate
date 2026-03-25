package service

import "context"

type Notifier interface {
	SendMessage(ctx context.Context, msg string) error
}
