package main

import (
	"context"
	"fmt"
	"log"
	"order/pkg/app"
	"order/pkg/db"
	"os"
	"os/signal"
	"platform/pkg/logger"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	ctx := context.Background()
	config, err := initConfig()
	if err != nil {
		log.Printf("failed to init config: %v\n", err)
	}
	config.ReadHeaderTimeout = readHeaderTimeout
	config.ShutdownTimeout = shutdownTimeout
	database, err := db.NewDB(ctx, config.DbUri)
	db.Migrate(ctx, database, config.MigrationsDir)

	if err != nil {
		log.Fatal(err)
	}

	var a *app.App
	a, err = app.New(ctx, config)
	if err != nil {
		logger.Error(ctx, "❌ Ошибка при создании приложения", zap.Error(err))
		return
	}

	err = a.Run(ctx, config)
	if err != nil {
		logger.Error(ctx, "❌ Ошибка при работе приложения", zap.Error(err))
		return
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Завершение работы сервера...")

	a.Stop()
}

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	secretsMapping := map[string]*string{
		"DB_URI":                   &config.DbUri,
		"MIGRATIONS_DIR":           &config.MigrationsDir,
		"HTTP_PORT":                &config.HttpPort,
		"SERVER_INVENTORY_ADDRESS": &config.ServerInventoryAddress,
		"SERVER_PAYMENT_ADDRESS":   &config.ServerPaymentAddress,
		"ORDER_KAFKA_BROKER":       &config.KafkaBroker,
		"CONSUME_TOPIC_NAME":       &config.ConsumeTopicName,
		"PRODUCE_TOPIC_NAME":       &config.ProduceTopicName,
		"ORDER_CONSUMER_GROUP_ID":  &config.ConsumerGroupId,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	return &config, nil
}
