package main

import (
	"log"
	"order/pkg/app"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	a := app.New(&app.Config{Port: httpPort, ReadHeaderTimeout: readHeaderTimeout, ShutdownTimeout: shutdownTimeout}, serverInventoryAddress, serverPaymentAddress)
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	a.Stop()
}
