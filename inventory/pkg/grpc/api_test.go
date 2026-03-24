package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"inventory/pkg/db"
	"inventory/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"
)

const bufSize = 1024 * 1024

var (
	testMongoC     *mongodb.MongoDBContainer
	testDB         *db.DB
	testRepo       *db.Repository
	testSvc        service.InventoryService
	testApi        *Api
	testGRPCServer *grpc.Server
	testClientConn *grpc.ClientConn
	testCtx        = context.Background()
	testCleanup    func()

	seedUUID1 = uuid.New().String()
	seedUUID2 = uuid.New().String()
	seedUUID3 = uuid.New().String()
	seedUUID4 = uuid.New().String()
	seedUUID5 = uuid.New().String()
)

func TestMain(m *testing.M) {
	var err error

	// 1. Запускаем MongoDB контейнер
	testMongoC, err = mongodb.Run(testCtx,
		"mongo:7",
		mongodb.WithUsername("testuser"),
		mongodb.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").
				WithOccurrence(1).
				WithStartupTimeout(40*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start mongodb: %v", err)
	}

	connStr, err := testMongoC.ConnectionString(testCtx)
	if err != nil {
		log.Fatalf("failed to get mongo conn string: %v", err)
	}

	// 2. Подключаемся к MongoDB
	client, err := mongo.Connect(testCtx, options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}
	if err = client.Ping(testCtx, nil); err != nil {
		log.Fatalf("mongo ping failed: %v", err)
	}

	coll := client.Database("testinv").Collection("part")
	testDB = &db.DB{
		MongoClient:     client,
		MongoCollection: coll,
	}

	// 3. Создаём репозиторий → сервис → api
	testRepo = db.NewRepository(testDB)
	testSvc = service.NewService(testRepo)
	testApi = New(testSvc)

	// 4. Запускаем in-memory gRPC сервер с bufconn
	testGRPCServer = grpc.NewServer()
	inventoryV1.RegisterInventoryServiceServer(testGRPCServer, testApi)

	listener := bufconn.Listen(bufSize)

	// запускаем сервер в отдельной горутине
	go func() {
		if err := testGRPCServer.Serve(listener); err != nil {
			log.Fatalf("grpc serve exited with error: %v", err)
		}
	}()

	// 5. Создаём клиентское соединение через bufconn (без реального TCP)
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}

	// Добавляем passthrough:///, чтобы отключить DNS lookup
	testClientConn, err = grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create grpc client conn: %v", err)
	}

	// Cleanup
	testCleanup = func() {
		testClientConn.Close()
		testGRPCServer.GracefulStop()
		client.Disconnect(testCtx)
		testcontainers.TerminateContainer(testMongoC)
	}

	// подготовка данных
	createParts()

	// Запуск тестов
	code := m.Run()
	testCleanup()
	os.Exit(code)
}

func createParts() {
	now := time.Now()
	seedParts := []interface{}{
		bson.M{
			"uuid":           seedUUID1,
			"name":           "Шарикоподшипник 6204-2RS",
			"description":    "Радиальный однорядный шарикоподшипник",
			"price":          350.00,
			"stock_quantity": 500,
			"manufacturer":   bson.M{"name": "Bosch", "country": "Germany", "website": ""},
			"tags":           []string{"metal", "mechanics", "gear"},
			"category":       0,
			"created_at":     now,
		},
		bson.M{
			"uuid":           seedUUID2,
			"name":           "Датчик температуры PT100",
			"description":    "Промышленный термометр",
			"price":          2150.75,
			"stock_quantity": 45,
			"manufacturer":   bson.M{"name": "test name", "country": "Moscow", "website": "https://moscow.com"},
			"tags":           []string{"fuel", "Moscow"},
			"category":       1,
			"created_at":     now,
		},
		bson.M{
			"uuid":           seedUUID3,
			"name":           "Пневмоцилиндр ISO 15552",
			"description":    "Стандартный пневматический цилиндр",
			"price":          4500.00,
			"stock_quantity": 15,
			"manufacturer":   bson.M{"name": "test name", "country": "Rostov", "website": "https://rostov.com"},
			"tags":           []string{"engine", "Rostov"},
			"category":       2,
			"created_at":     now,
		},
		bson.M{
			"uuid":           seedUUID4,
			"name":           "Болт М8х30 DIN 912",
			"description":    "Винт с внутренним шестигранником",
			"price":          15.50,
			"stock_quantity": 2000,
			"manufacturer":   bson.M{"name": "test name", "country": "Rostov", "website": "https://rostov.com"},
			"tags":           []string{"Rostov"},
			"category":       0,
			"created_at":     now,
		},
		bson.M{
			"uuid":           seedUUID5,
			"name":           "Гидравлический клапан",
			"description":    "Предохранительный клапан",
			"price":          12400.00,
			"stock_quantity": 8,
			"category":       0,
			"manufacturer":   bson.M{"name": "test name", "country": "Moscow", "website": "https://moscow.com"},
			"tags":           []string{"engine", "Moscow"},
			"created_at":     now,
		},
	}

	if _, err := testDB.MongoCollection.InsertMany(testCtx, seedParts); err != nil {
		log.Fatalf("failed to seed database in TestMain: %v", err)
	}
}

func TestGetPartSuccess(t *testing.T) {
	ctx := context.Background()

	// Создаём запрос
	req := &inventoryV1.GetPartRequest{Uuid: seedUUID1}

	// Вызываем метод через динамический клиент
	var resp inventoryV1.GetPartResponse
	err := testClientConn.Invoke(
		ctx,
		"/inventory.v1.InventoryService/GetPart",
		req,
		&resp,
	)
	require.NoError(t, err, "Invoke GetPart failed")

	// Проверяем ответ
	require.NotNil(t, resp.Part, "Part should be returned")
	assert.Equal(t, seedUUID1, resp.Part.Uuid)
	assert.Equal(t, "Шарикоподшипник 6204-2RS", resp.Part.Name)
	assert.Equal(t, 350.00, resp.Part.Price)
}

func TestGetPartNotFound(t *testing.T) {
	ctx := context.Background()

	req := &inventoryV1.GetPartRequest{Uuid: uuid.New().String()}

	var resp inventoryV1.GetPartResponse
	err := testClientConn.Invoke(
		ctx,
		"/inventory.v1.InventoryService/GetPart",
		req,
		&resp,
	)

	require.Error(t, err, "expected error for missing part")
	assert.Contains(t, err.Error(), "can't get part")
	assert.Empty(t, resp.Part)
}

func TestGetListParts(t *testing.T) {
	ctx := context.Background()

	// структура тестовых кейсов
	tests := []struct {
		name          string                   // Название подтеста
		filter        *inventoryV1.PartsFilter // Входные данные (фильтр)
		expectedLen   int                      // Ожидаемое количество деталей
		expectedUUIDs []string                 // Ожидаемые UUID в ответе

		// Опциональные поля для проверки конкретной детали
		checkPartUUID string
		expectedName  string
		expectedPrice float64
	}{
		{
			name:          "Without filter",
			filter:        &inventoryV1.PartsFilter{},
			expectedLen:   5,
			expectedUUIDs: []string{seedUUID1, seedUUID2, seedUUID3, seedUUID4, seedUUID5},
			checkPartUUID: seedUUID4,
			expectedName:  "Болт М8х30 DIN 912",
			expectedPrice: 15.50,
		},
		{
			name: "Filter by multiple UUIDs",
			filter: &inventoryV1.PartsFilter{
				Uuids: []string{seedUUID1, seedUUID2, seedUUID3},
			},
			expectedLen:   3,
			expectedUUIDs: []string{seedUUID1, seedUUID2, seedUUID3},
			checkPartUUID: seedUUID1,
			expectedName:  "Шарикоподшипник 6204-2RS",
			expectedPrice: 350.00,
		},
		{
			name: "Filter by exact Name",
			filter: &inventoryV1.PartsFilter{
				Names: []string{"Гидравлический клапан"},
			},
			expectedLen:   1,
			expectedUUIDs: []string{seedUUID5},
			checkPartUUID: seedUUID5,
			expectedName:  "Гидравлический клапан",
			expectedPrice: 12400.00,
		},
		{
			name: "Filter by exact Categories",
			filter: &inventoryV1.PartsFilter{
				Categories: []inventoryV1.Category{
					inventoryV1.Category_CATEGORY_ENGINE,
					inventoryV1.Category_CATEGORY_FUEL,
				},
			},
			expectedLen:   2,
			expectedUUIDs: []string{seedUUID2, seedUUID3},
			checkPartUUID: seedUUID2,
			expectedName:  "Датчик температуры PT100",
			expectedPrice: 2150.75,
		},
		{
			name: "Filter by exact Tags",
			filter: &inventoryV1.PartsFilter{
				Tags: []string{"engine", "Rostov"},
			},
			expectedLen:   1,
			expectedUUIDs: []string{seedUUID3},
			checkPartUUID: seedUUID3,
			expectedName:  "Пневмоцилиндр ISO 15552",
			expectedPrice: 4500.00,
		},
		{
			name: "Filter with empty result",
			filter: &inventoryV1.PartsFilter{
				Tags: []string{"engine", "Rost"},
			},
			expectedLen:   0,
			expectedUUIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &inventoryV1.GetListPartsRequest{
				Filter: tt.filter,
			}

			var resp inventoryV1.GetListPartsResponse
			err := testClientConn.Invoke(
				ctx,
				"/inventory.v1.InventoryService/GetListParts",
				req,
				&resp,
			)

			// Базовые проверки, общие для всех
			require.NoError(t, err, "Invoke GetListParts failed")
			assert.Len(t, resp.Parts, tt.expectedLen)

			// Проверка списка UUID
			var returnedUUIDs []string
			for _, part := range resp.Parts {
				returnedUUIDs = append(returnedUUIDs, part.Uuid)
			}
			assert.ElementsMatch(t, tt.expectedUUIDs, returnedUUIDs, "список возвращенных UUID не совпадает")

			// Точечная проверка конкретной детали
			if tt.checkPartUUID != "" {
				var foundPart *inventoryV1.Part
				for _, part := range resp.Parts {
					if part.Uuid == tt.checkPartUUID {
						foundPart = part
						break
					}
				}
				require.NotNil(t, foundPart, "в ответе должна присутствовать деталь с заданным UUID")
				assert.Equal(t, tt.expectedName, foundPart.Name)
				assert.Equal(t, tt.expectedPrice, foundPart.Price)
			}
		})
	}
}
