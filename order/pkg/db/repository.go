package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"log"
	"os"
	"sync"
)

type OrderRepository interface {
	CreateOrder(order *Order)
	GetOrder(orderUuid string) (*Order, error)
	UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error
}

type Repository struct {
	mu   sync.RWMutex
	pool *pgxpool.Pool
}

var _ OrderRepository = (*Repository)(nil)

func NewRepository() *Repository {
	dbURI := os.Getenv("DB_URI")

	// Создаем пул соединений с базой данных
	pool, err := pgxpool.New(context.Background(), dbURI)
	if err != nil {
		log.Printf("failed to connect to database: %v\n", err)
		return &Repository{}
	}
	log.Printf("Connecting to database with URI: %s", dbURI)

	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateOrder(order *Order) {
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
		return
	}

	var orderId string
	err = r.pool.QueryRow(context.Background(), query, args...).Scan(&orderId)
	if err != nil {
		log.Printf("failed to insert note: %v\n", err)
		return
	}

	log.Printf("inserted note with id: %s\n", orderId)
}

func (r *Repository) GetOrder(orderUuid string) (*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	row, err := r.pool.Query(context.Background(),
		`SELECT * FROM "order" WHERE order_uuid = $1 LIMIT 1`,
		orderUuid,
	)
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

func (r *Repository) UpdateOrder(orderUuid string, transactionUuid *string, status *OrderStatus, paymentMethod *OrderPaymentMethod) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.GetOrder(orderUuid)

	if ok != nil {
		return fmt.Errorf("order with id %s not found", orderUuid)
	}

	builderUpdate := sq.Update("\"order\"").
		PlaceholderFormat(sq.Dollar).
		Set("updated_at", time.Now())

	if transactionUuid != nil {
		builderUpdate = builderUpdate.Set("transaction_uuid", *transactionUuid)
	}

	if status != nil {
		builderUpdate = builderUpdate.Set("status", status.String())
	}

	if paymentMethod != nil {
		builderUpdate = builderUpdate.Set("payment_method", paymentMethod.String())
	}

	builderUpdate = builderUpdate.Where(sq.Eq{"order_uuid": orderUuid})
	query, args, err := builderUpdate.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v\n", err)
		return fmt.Errorf("failed to build query: %v\n", err)
	}

	res, err := r.pool.Exec(context.Background(), query, args...)
	if err != nil {
		log.Printf("failed to update order: %v\n", err)
		return fmt.Errorf("failed to update order: %v\n", err)
	}

	log.Printf("updated %d rows", res.RowsAffected())

	return nil
}
