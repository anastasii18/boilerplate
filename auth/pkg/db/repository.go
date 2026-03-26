package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"platform/migrator"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthRepository interface {
	Register(ctx context.Context, user *User) (*string, error)
	Login(ctx context.Context, user *User) (*string, error)
	GetUserByUuid(ctx context.Context, userUuid string) (*User, error)
}

type Repository struct {
	db *DB
}

type DB struct {
	*pgxpool.Pool
}

func NewDB(ctx context.Context, dbURI string) (*DB, error) {
	// Создаем пул соединений с базой данных
	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connecting to database with URI: %s", dbURI)

	return &DB{pool}, nil
}

func Migrate(ctx context.Context, db *DB, migrationsDir string) {
	// Инициализируем мигратор
	dbConfig := db.Pool.Config().ConnConfig
	sqlDB := stdlib.OpenDB(*dbConfig)

	migratorRunner := migrator.NewMigrator(sqlDB, migrationsDir)

	err := migratorRunner.Up()
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v\n", err)
		return
	}
}

var _ AuthRepository = (*Repository)(nil)

func NewRepository(database *DB) *Repository {
	return &Repository{
		db: database,
	}
}

func (r *Repository) Register(ctx context.Context, user *User) (*string, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	builderInsert := sq.Insert("users").
		PlaceholderFormat(sq.Dollar).
		Columns("user_uuid", "login", "email", "hashed_password", "notification_methods").
		Values(user.UserUuid, user.Login, user.Email, string(password), user.NotificationMethods).
		Suffix("RETURNING user_uuid")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v\n", err)
		return nil, err
	}

	var userUuid *string
	err = r.db.QueryRow(ctx, query, args...).Scan(&userUuid)
	if err != nil {
		log.Fatalf("failed to insert user: %v\n", err)
		return userUuid, err
	}

	log.Printf("inserted user with id: %s\n", userUuid)

	return userUuid, nil
}

func (r *Repository) GetUserByUuid(ctx context.Context, userUuid string) (*User, error) {
	builderSelect := sq.Select("*").
		From("users").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"user_uuid": userUuid}).
		Limit(1)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer row.Close()

	user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", userUuid)
		}
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	builderSelect := sq.Select("*").
		From("users").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"login": login}).
		Limit(1)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer row.Close()

	user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[User])

	return &user, err
}

func (r *Repository) Login(ctx context.Context, user *User) (*string, error) {
	dbUser, err := r.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// сохранять в redis
	session_uuid := uuid.New().String()
	return Ptr(session_uuid), nil
}

func Ptr[T comparable](t T) *T {
	var def T
	if t == def {
		return nil
	}
	return &t
}
