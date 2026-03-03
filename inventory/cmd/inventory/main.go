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
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const grpcPort = 50051

func main() {
	seed := flag.Bool("seed", false, "заполнить тестовыми данными")
	initIndexes := flag.Bool("init-indexes", false, "создать/обновить индексы в MongoDB")
	flag.Parse()
	ctx := context.Background()
	mongoDB, err := mongoInit(ctx, initIndexes)
	if err != nil {
		log.Fatal(err)
	}
	// Регистрируем наш сервис
	a := app.New(ctx, grpcPort, mongoDB, seed)
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

	if err := mongoDB.MongoClient.Disconnect(ctx); err != nil {
		log.Printf("⚠️ Ошибка при отключении MongoDB: %v", err)
	} else {
		log.Println("✅ MongoDB disconnected")
	}

	log.Println("✅ Server stopped")
}

func mongoInit(ctx context.Context, initIndexes *bool) (*db.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env: %v", err)
		// можно не возвращать ошибку, если .env необязателен
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGO_URI не задан")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Проверка соединения
	if err := client.Ping(ctx, nil); err != nil {
		err := client.Disconnect(ctx)
		if err != nil {
			return nil, err
		} // сразу закрываем, если пинг не прошёл
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	dbName := os.Getenv("MONGO_INITDB_DATABASE")

	database := client.Database(dbName)
	collection := database.Collection("part")

	if app.Val(initIndexes) {
		// Создание индексов
		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "uuid", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("uuid_unique"),
			},
		}

		opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
		names, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
		if err != nil {
			log.Printf("❌ Ошибка создания индекса uuid_unique: %v", err)
		} else {
			log.Printf("✅ Индексы созданы: %v", names)
		}
	}

	return &db.DB{MongoClient: client, MongoCollection: collection}, nil
}
