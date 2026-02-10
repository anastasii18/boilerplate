package main

import (
	"inventory/pkg/app"
	"inventory/pkg/db"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const grpcPort = 50051

func main() {
	// Регистрируем наш сервис
	a := app.New(grpcPort, db.NewDB())
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.Server.GracefulStop()
	log.Println("✅ Server stopped")
}
