package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"log"
	"sync"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, orderUuid string) (*Order, error)
	UpdateOrder(ctx context.Context, orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

type Repository struct {
	mu sync.RWMutex
	DB *DB
}

type DB struct {
	*pgxpool.Pool
}

func NewDB(ctx context.Context) (*DB, error) {
	dbURI := os.Getenv("DB_URI")

	// Создаем пул соединений с базой данных
	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v\n", err)
	}
	log.Printf("Connecting to database with URI: %s", dbURI)
	return &DB{pool}, nil
}

var _ OrderRepository = (*Repository)(nil)

func NewRepository(database *DB) *Repository {
	return &Repository{
		DB: database,
	}
}

func (r *Repository) CreateOrder(ctx context.Context, order *Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	builderInsert := sq.Insert("\"order\"").
		PlaceholderFormat(sq.Dollar).
		Columns("order_uuid", "transaction_uuid", "status", "payment_method",
			"part_uuids", "user_uuid", "total_price", "created_at", "updated_at").
		Values(order.OrderUuid, order.TransactionUuid, order.Status, order.PaymentMethod,
			order.PartUuids, order.UserUuid, order.TotalPrice, order.CreatedAt, order.UpdatedAt).
		Suffix("RETURNING order_uuid")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v\n", err)
		return err
	}

	var orderId string
	err = r.DB.QueryRow(ctx, query, args...).Scan(&orderId)
	if err != nil {
		log.Printf("failed to insert note: %v\n", err)
		return err
	}

	log.Printf("inserted note with id: %s\n", orderId)

	return nil
}

func (r *Repository) GetOrder(ctx context.Context, orderUuid string) (*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	builderSelect := sq.Select("*").
		From("\"order\"").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"order_uuid": orderUuid}).
		Limit(1)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer row.Close()

	order, err := pgx.CollectOneRow(row, pgx.RowToStructByName[Order])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %s", orderUuid)
		}
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	return &order, nil
}

func (r *Repository) UpdateOrder(ctx context.Context, orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.GetOrder(ctx, orderUuid)

	if ok != nil {
		return fmt.Errorf("order with id %s not found", orderUuid)
	}

	builderUpdate := sq.Update("\"order\"").
		PlaceholderFormat(sq.Dollar).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"order_uuid": orderUuid})

	if transactionUuid != nil {
		builderUpdate = builderUpdate.Set("transaction_uuid", *transactionUuid)
	}

	if status != nil {
		builderUpdate = builderUpdate.Set("status", status.String())
	}

	if paymentMethod != nil {
		builderUpdate = builderUpdate.Set("payment_method", paymentMethod.String())
	}

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v\n", err)
		return fmt.Errorf("failed to build query: %v\n", err)
	}

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to update order: %v\n", err)
		return fmt.Errorf("failed to update order: %v\n", err)
	}

	log.Printf("updated %d rows", res.RowsAffected())

	return nil
}
