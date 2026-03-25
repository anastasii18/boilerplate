package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InventoryRepository interface {
	GetParts(ctx context.Context, filter PartSearch) (map[string]*Part, error)
	GetPart(ctx context.Context, id string) (*Part, error)
}

type Repository struct {
	db *DB
}

var _ InventoryRepository = (*Repository)(nil)

func NewRepository(db *DB) *Repository {
	return &Repository{
		db: db,
	}
}

func NewDB(ctx context.Context, uri, dbName string) (*DB, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Проверка соединения
	if err = client.Ping(ctx, nil); err != nil {
		err = client.Disconnect(ctx)
		if err != nil {
			return nil, err
		} // сразу закрываем, если пинг не прошёл
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	collection := database.Collection("part")

	return &DB{MongoClient: client, MongoCollection: collection}, nil
}

func (db *DB) InitIndex(ctx context.Context, initIndexes bool) error {
	if !initIndexes {
		return nil
	}
	// Создание индексов
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uuid_unique"),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	names, err := db.MongoCollection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		return fmt.Errorf("ошибка создания индекса uuid_unique: %w", err)
	}

	log.Printf("Индексы созданы: %v", names)
	return nil
}

type DB struct {
	MongoClient     *mongo.Client
	MongoCollection *mongo.Collection
}

// Seed наполняет базу тестовыми данными
func (r *Repository) Seed(ctx context.Context) *Repository {
	now := time.Now()

	parts := []Part{
		{
			Uuid:          "4ea110c1-342e-4102-a0e6-ae0d9a04ce13",
			Name:          "Шестерня",
			Description:   "Металлическая шестерня для станка",
			Price:         1499.99,
			StockQuantity: 50,
			Dimensions: &Dimensions{
				Length: 12.5,
				Width:  8.0,
				Height: 3.2,
			},
			Manufacturer: &Manufacturer{
				Name:    "Bosch",
				Country: "Germany",
			},
			Tags:      []string{"metal", "mechanics", "gear"},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			Uuid:          "fbb05498-4db6-48c8-b945-3e56f4e5ad04",
			Name:          "test name",
			Description:   "test description",
			Price:         112.33,
			StockQuantity: 38,
			Category:      CATEGORY_FUEL,
			Manufacturer: &Manufacturer{
				Name:    "test name",
				Country: "Moscow",
				Website: "https://moscow.com",
			},
			Tags:      []string{"fuel", "Moscow"},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			Uuid:          "bf802b57-1c7d-41ff-9cb7-ee43dbadbf98",
			Name:          "two two",
			Description:   "test description",
			Price:         45.45,
			StockQuantity: 7,
			Category:      CATEGORY_ENGINE,
			Manufacturer: &Manufacturer{
				Name:    "test name",
				Country: "Rostov",
				Website: "https://rostov.com",
			},
			Tags:      []string{"engine", "Rostov"},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			Uuid:          "29a9ab94-c814-4828-9a02-b96598dbe299",
			Name:          "three three",
			Description:   "test description",
			Price:         66.77,
			StockQuantity: 90,
			Category:      CATEGORY_ENGINE,
			Manufacturer: &Manufacturer{
				Name:    "test name",
				Country: "Moscow",
				Website: "https://moscow.com",
			},
			Tags:      []string{"engine", "Moscow"},
			CreatedAt: now,
			UpdatedAt: &now,
		},
	}

	insertedCount := 0
	for _, part := range parts {
		filter := bson.M{"uuid": part.Uuid}
		count, err := r.db.MongoCollection.CountDocuments(ctx, filter)
		if err != nil {
			log.Printf("Ошибка проверки uuid %s: %v", part.Uuid, err)
			continue
		}
		if count > 0 {
			log.Printf("Uuid %s уже существует, пропускаем", part.Uuid)
			continue
		}

		_, err = r.db.MongoCollection.InsertOne(ctx, part)
		if err != nil {
			log.Printf("Ошибка вставки uuid %s: %v", part.Uuid, err)
			continue
		}

		insertedCount++
	}

	log.Printf("Вставлено %d новых деталей", insertedCount)
	return r
}

func (r *Repository) GetParts(ctx context.Context, filter PartSearch) (map[string]*Part, error) {
	cursor, err := r.db.MongoCollection.Find(ctx, NewInventoryFilter(filter))
	if err != nil {
		return nil, fmt.Errorf("ошибка Find: %w", err)
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			log.Printf("failed to close mongo cursor: %v", closeErr)
		}
	}()

	// Собираем результат в map
	result := make(map[string]*Part)

	for cursor.Next(ctx) {
		var part Part
		if err := cursor.Decode(&part); err != nil {
			return nil, fmt.Errorf("ошибка декодирования: %w", err)
		}
		result[part.Uuid] = &part
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("ошибка курсора: %w", err)
	}

	return result, nil
}

func NewInventoryFilter(filter PartSearch) bson.M {
	mongoFilter := bson.M{}

	if filter.Uuids != nil {
		mongoFilter["uuid"] = bson.M{"$in": filter.Uuids}
	}

	if filter.Names != nil {
		mongoFilter["name"] = bson.M{"$in": filter.Names}
	}

	if filter.Categories != nil {
		mongoFilter["category"] = bson.M{"$in": filter.Categories}
	}

	if filter.ManufacturerCountries != nil {
		mongoFilter["manufacturer.country"] = bson.M{"$in": filter.ManufacturerCountries}
	}

	if filter.Tags != nil {
		mongoFilter["tags"] = bson.M{"$all": filter.Tags}
	}

	return mongoFilter
}

func (r *Repository) GetPart(ctx context.Context, id string) (*Part, error) {
	var part Part
	err := r.db.MongoCollection.FindOne(ctx, bson.M{"uuid": id}).Decode(&part)
	if err != nil {
		return nil, fmt.Errorf("part with id %s not found", id)
	}

	return &part, nil
}
