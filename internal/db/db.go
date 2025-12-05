package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(dsn string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to parse DB DSN: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	return pool
}

