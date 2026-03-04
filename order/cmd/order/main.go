package main

import (
	"context"
	"fmt"
	"log"
	"order/pkg/app"
	"order/pkg/db"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
	cfg := app.Config{Port: config.HttpPort, ReadHeaderTimeout: readHeaderTimeout, ShutdownTimeout: shutdownTimeout}
	database, err := db.NewDB(ctx, config.DbUri)
	db.Migrate(ctx, database, config.MigrationsDir)

	if err != nil {
		log.Fatal(err)
	}
	a := app.New(ctx, &cfg, config.ServerInventoryAddress, config.ServerPaymentAddress, database)
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	a.Stop()
}

type Config struct {
	DbUri                  string
	MigrationsDir          string
	HttpPort               string
	ServerInventoryAddress string
	ServerPaymentAddress   string
}

func initConfig() (*Config, error) {
	var config Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %v", err)
	}

	secretsMapping := map[string]*string{
		"DB_URI":                   &config.DbUri,
		"MIGRATIONS_DIR":           &config.MigrationsDir,
		"HTTP_PORT":                &config.HttpPort,
		"SERVER_INVENTORY_ADDRESS": &config.ServerInventoryAddress,
		"SERVER_PAYMENT_ADDRESS":   &config.ServerPaymentAddress,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	return &config, nil
}
