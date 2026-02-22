package main

import (
	"log"
	"os"
	"os/signal"
	"payment/pkg/app"
	"syscall"
)

const grpcPort = 50052

func main() {
	a := app.New(&app.Config{Port: grpcPort})
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.Server.GracefulStop()
	log.Println("✅ Server stopped")
}
