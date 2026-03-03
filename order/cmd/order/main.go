package main

import (
	"context"
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
	httpPort               = "8080"
	serverInventoryAddress = "localhost:50051"
	serverPaymentAddress   = "localhost:50052"
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed to load .env file: %v\n", err)
		return
	}
	cfg := app.Config{Port: httpPort, ReadHeaderTimeout: readHeaderTimeout, ShutdownTimeout: shutdownTimeout}
	database, err := db.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	a := app.New(ctx, &cfg, serverInventoryAddress, serverPaymentAddress, database)
	a.Start()
	a.Migrate(context.Background())

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	a.Stop()
}
