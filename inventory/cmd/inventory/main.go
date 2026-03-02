package main

import (
	"context"
	"inventory/pkg/app"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const grpcPort = 50051

func main() {
	// Регистрируем наш сервис
	a, err := app.New(grpcPort)
	if err != nil {
		log.Fatalf("не удалось открыть порт: %v", err)
	}
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.Server.GracefulStop()
	// Закрываем соединение с MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.MongoClient.Disconnect(ctx); err != nil {
		log.Printf("⚠️ Ошибка при отключении MongoDB: %v", err)
	} else {
		log.Println("✅ MongoDB disconnected")
	}

	log.Println("✅ Server stopped")
}
