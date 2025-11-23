package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

func NewDB(ctx context.Context, connString string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Failed to parse db conf: %v", err)
	}

	// Настройки пула, минимальное и максимальное кол-во соединений
	cfg.MinConns = 1
	cfg.MaxConns = 10

	// Создаем пул
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create db pool: %v", err)
	}

	// Проверяем соединение
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxPing); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	log.Println("Successfully connected to db")

	return pool
}
