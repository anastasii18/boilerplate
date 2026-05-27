package main

import (
	"assembly/pkg/app"
	"context"
	"fmt"
	"os"
	"os/signal"
	"platform/pkg/logger"
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
		logger.Error(appCtx, "Не удалось создать конфиг", zap.Error(err))
		return
	}

	a, err := app.New(appCtx, config)
	if err != nil {
		logger.Error(appCtx, "Не удалось создать приложение", zap.Error(err))
		return
	}

	err = a.Run(appCtx)
	if err != nil {
		logger.Error(appCtx, "Ошибка при работе приложения", zap.Error(err))
		return
	}
}

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	secretsMapping := map[string]*string{
		"ORDER_KAFKA_BROKER":      &config.KafkaBroker,
		"CONSUME_TOPIC_NAME":      &config.ConsumeTopicName,
		"PRODUCE_TOPIC_NAME":      &config.ProduceTopicName,
		"ORDER_CONSUMER_GROUP_ID": &config.ConsumerGroupId,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	return &config, nil
}

func gracefulShutdown() {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
}
