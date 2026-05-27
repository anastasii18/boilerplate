package main

import (
	"context"
	"fmt"
	"log"
	"notification/pkg/app"
	"os"
	"os/signal"
	"platform/pkg/logger"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()

	config, err := initConfig()
	if err != nil {
		log.Printf("failed to init config: %v\n", err)
		return
	}
	a, err := app.New(appCtx, config)
	if err != nil {
		logger.Error(appCtx, "Не удалось создать приложение", zap.Error(err))
		return
	}

	err = a.Run(appCtx)
	if err != nil {
		logger.Error(appCtx, "Не удалось запустить приложение", zap.Error(err))
		return
	}

	log.Println("Получен сигнал завершения. Выходим...")
}

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	secretsMapping := map[string]*string{
		"ORDER_KAFKA_BROKER":      &config.KafkaBroker,
		"ORDER_CONSUMER_GROUP_ID": &config.GroupId,
		"TELEGRAM_BOT_TOKEN":      &config.TelegramBotToken,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	// Достаем топики
	topicsStr := os.Getenv("KAFKA_TOPICS")
	if topicsStr != "" {
		config.TopicNames = strings.Split(topicsStr, ",")
	}

	return &config, nil
}
