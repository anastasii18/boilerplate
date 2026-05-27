package app

import (
	"context"
	"platform/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type App struct {
	config      *Config
	diContainer *diContainer
}

type Config struct {
	KafkaBroker      string
	TopicNames       []string
	GroupId          string
	TelegramBotToken string
}

func New(ctx context.Context, config *Config) (*App, error) {
	a := &App{config: config}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDI,
		a.initLogger,
		a.initTelegramBot,
		a.initConsumer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) initDI(_ context.Context) error {
	app.diContainer = NewDiContainer()
	return nil
}

// Уровень логирования (debug, info, warn, error)
func (app *App) initLogger(ctx context.Context) error {
	return logger.Init(
		"debug",
		true,
	)
}

func (app *App) initTelegramBot(ctx context.Context) error {
	// Получаем бота из DI контейнера
	telegramBot, err := app.diContainer.TelegramBot(ctx, app.config.TelegramBotToken)
	if err != nil {
		return err
	}

	// Регистрируем обработчик для активации бота
	telegramBot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		logger.Info(ctx, "chat id", zap.Int64("chat_id", update.Message.Chat.ID))

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Notification Bot активирован! Теперь вы будете получать уведомления об успешной сборке и оплате заказов.",
		})
		if err != nil {
			logger.Error(ctx, "Failed to send activation message", zap.Error(err))
		}
	})

	// Запускаем бота в фоне
	go func() {
		logger.Info(ctx, "Telegram bot started...")
		telegramBot.Start(ctx)
	}()

	return nil
}

// Инициализация и запуск консьюмера
func (app *App) initConsumer(ctx context.Context) error {
	// Достаем консьюмер из DI-контейнера
	consumerService, err := app.diContainer.NotificationConsumer(ctx, app.config.KafkaBroker, app.config.GroupId, app.config.TelegramBotToken, app.config.TopicNames)
	if err != nil {
		return err
	}

	// Запускаем чтение сообщений из Kafka
	go func() {
		logger.Info(ctx, "Запуск Kafka Consumer для notification...")

		err := consumerService.RunConsumer(ctx)
		if err != nil {
			logger.Error(ctx, "Ошибка во время работы Kafka Consumer", zap.Error(err))
		}
	}()

	return nil
}
