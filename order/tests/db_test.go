package tests

import (
	"context"
	"log"
	"order/pkg/db"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"testing"
	"time"
)

var (
	testDB      *db.DB
	testRepo    *db.Repository
	testConnStr string
	testPgC     *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error

	// 1. Запускаем контейнер
	testPgC, err = postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(45*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}

	// Получаем строку подключения
	testConnStr, err = testPgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// 2. Создаём пул соединений
	pool, err := pgxpool.New(ctx, testConnStr)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	testDB = &db.DB{Pool: pool}

	// 3. Применяем миграции
	migrationsDir := filepath.Join("../pkg/db/", "migrations")
	db.Migrate(ctx, testDB, migrationsDir)

	// 4. Создаём репозиторий
	testRepo = db.NewRepository(testDB)

	// Запускаем все тесты
	code := m.Run()

	// 5. Cleanup после всех тестов
	pool.Close()
	if err := testcontainers.TerminateContainer(testPgC); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func TestCreateAndGetOrder(t *testing.T) {
	ctx := context.Background()

	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	partUUIDs := []string{uuid.New().String(), uuid.New().String()}

	original := &db.Order{
		OrderUuid:       orderUUID,
		TransactionUuid: nil,
		Status:          db.PENDING_PAYMENT,
		PaymentMethod:   db.UNKNOWN,
		PartUuids:       partUUIDs,
		UserUuid:        userUUID,
		TotalPrice:      1499.95,
		CreatedAt:       time.Now().Truncate(time.Second),
	}

	// создаём заказ
	err := testRepo.CreateOrder(ctx, original)
	require.NoError(t, err, "CreateOrder failed")

	// читаем
	found, err := testRepo.GetOrder(ctx, orderUUID)
	require.NoError(t, err, "GetOrder failed")
	require.NotNil(t, found)

	assert.Equal(t, original.OrderUuid, found.OrderUuid)
	assert.Equal(t, original.UserUuid, found.UserUuid)
	assert.Equal(t, original.TotalPrice, found.TotalPrice)
	assert.Equal(t, original.Status, found.Status)
	assert.Equal(t, original.PaymentMethod, found.PaymentMethod)
	assert.ElementsMatch(t, original.PartUuids, found.PartUuids)
	assert.WithinDuration(t, original.CreatedAt, found.CreatedAt, time.Second)
}

func TestCreateAndUpdateOrder(t *testing.T) {
	ctx := context.Background()

	orderUUID := uuid.New().String()
	userUUID := uuid.New().String()
	partUUIDs := []string{uuid.New().String(), uuid.New().String()}

	original := &db.Order{
		OrderUuid:       orderUUID,
		TransactionUuid: nil,
		Status:          db.PENDING_PAYMENT,
		PaymentMethod:   db.UNKNOWN,
		PartUuids:       partUUIDs,
		UserUuid:        userUUID,
		TotalPrice:      1499.95,
		CreatedAt:       time.Now().Truncate(time.Second),
	}

	// создаём заказ
	err := testRepo.CreateOrder(ctx, original)
	require.NoError(t, err, "CreateOrder failed")

	// обновляем
	transactionId := ptr(uuid.New().String())
	status := db.PAID
	paymentMethod := db.SBP
	err = testRepo.UpdateOrder(ctx, orderUUID, transactionId, ptr(status), ptr(paymentMethod))
	require.NoError(t, err, "GetOrder failed")

	// читаем
	found, err := testRepo.GetOrder(ctx, orderUUID)
	require.NoError(t, err, "GetOrder failed")
	require.NotNil(t, found)

	assert.Equal(t, transactionId, found.TransactionUuid)
	assert.Equal(t, status, found.Status)
	assert.Equal(t, paymentMethod, found.PaymentMethod)
	assert.WithinDuration(t, time.Now(), val(found.UpdatedAt), time.Second)
}

func ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}

func val[T any, P *T](p P) T {
	if p != nil {
		return *p
	}
	var def T
	return def
}
