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
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()
	defer gracefulShutdown()

	config, err := initConfig()
	if err != nil {
		log.Printf("failed to init config: %v\n", err)
	}
	_, err = app.New(appCtx, config)
	if err != nil {
		logger.Error(appCtx, "❌ Не удалось создать приложение", zap.Error(err))
		return
	}

	<-appCtx.Done()

	log.Println("Получен сигнал завершения. Выходим...")
}

func gracefulShutdown() {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
