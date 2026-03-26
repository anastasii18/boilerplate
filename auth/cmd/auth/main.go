package main

import (
	"auth/pkg/app"
	"auth/pkg/db"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	config, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	database, err := db.NewDB(ctx, config.DbUri)
	db.Migrate(ctx, database, config.MigrationsDir)

	if err != nil {
		log.Fatal(err)
	}

	// Регистрируем наш сервис
	a, err := app.New(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.DiContainer.Server.GracefulStop()

	log.Println("✅ Server stopped")
}

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	secretsMapping := map[string]*string{
		"DB_URI":         &config.DbUri,
		"GRPC_PORT":      &config.GrpcPort,
		"MIGRATIONS_DIR": &config.MigrationsDir,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	return &config, nil
}
