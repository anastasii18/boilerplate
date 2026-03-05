package main

import (
	"context"
	"flag"
	"fmt"
	"inventory/pkg/app"
	"inventory/pkg/db"
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

	mongoDB, err := db.NewDB(ctx, config.MongoURI, config.MongoDB)
	if err != nil {
		log.Fatal(err)
	}

	err = mongoDB.InitIndex(ctx, initIndexes)
	if err != nil {
		log.Fatal(err)
	}
	// Регистрируем наш сервис
	a := app.New(ctx, config, mongoDB, seed)
	a.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	a.Server.GracefulStop()

	// Закрываем соединение с MongoDB
	if err := mongoDB.MongoClient.Disconnect(ctx); err != nil {
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
