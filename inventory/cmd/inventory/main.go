package main

import (
	"context"
	"flag"
	"fmt"
	"inventory/pkg/app"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

var seed bool
var initIndexes bool

func main() {
	flag.BoolVar(&seed, "seed", false, "заполнить тестовыми данными")
	flag.BoolVar(&initIndexes, "init-indexes", false, "создать/обновить индексы в MongoDB")
	flag.Parse()

	config, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Регистрируем наш сервис
	a := app.New(ctx, config, initIndexes, seed)
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.Container.Server.GracefulStop()

	// Закрываем соединение с MongoDB
	if err := a.Container.DB.MongoClient.Disconnect(ctx); err != nil {
		log.Printf("⚠️ Ошибка при отключении MongoDB: %v", err)
	} else {
		log.Println("✅ MongoDB disconnected")
	}

	log.Println("✅ Server stopped")
}

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env: %v", err)
	}

	secretsMapping := map[string]*string{
		"MONGO_URI":             &config.MongoURI,
		"MONGO_INITDB_DATABASE": &config.MongoDB,
		"GRPC_PORT":             &config.GrpcPort,
	}
	for key, target := range secretsMapping {
		*target = os.Getenv(key)
	}

	return &config, nil
}
