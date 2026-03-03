package main

import (
	"context"
	"fmt"
	"log"
	"order/pkg/app"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	httpPort               = "8080"
	serverInventoryAddress = "localhost:50051"
	serverPaymentAddress   = "localhost:50052"
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed to load .env file: %v\n", err)
		return
	}
	cfg := app.Config{Port: httpPort, ReadHeaderTimeout: readHeaderTimeout, ShutdownTimeout: shutdownTimeout}
	pool, err := dbInit()
	if err != nil {
		log.Fatal(err)
	}
	a := app.New(&cfg, serverInventoryAddress, serverPaymentAddress, pool)
	a.Start()
	a.Migrate(context.Background())

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	a.Stop()
}

func dbInit() (*pgxpool.Pool, error) {
	dbURI := os.Getenv("DB_URI")

	// Создаем пул соединений с базой данных
	pool, err := pgxpool.New(context.Background(), dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v\n", err)
	}
	log.Printf("Connecting to database with URI: %s", dbURI)
	return pool, nil
}
