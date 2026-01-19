package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Trade is our "defined type" (a Go struct) that we will write/read.
type Trade struct {
	ID        int64
	Market    string
	Side      string
	Price     float64
	Quantity  int32
	CreatedAt time.Time
}

func main() {
	ctx := context.Background()

	// Use env var if set, otherwise default to our Docker local connection string.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// 1) Create table (idempotent)
	if err := createSchema(ctx, pool); err != nil {
		log.Fatalf("schema: %v", err)
	}

	// 2) Insert a Trade
	newTrade := Trade{
		Market:   "KALSHI:RATECUT-MAR",
		Side:     "YES",
		Price:    0.57,
		Quantity: 10,
	}

	inserted, err := insertTrade(ctx, pool, newTrade)
	if err != nil {
		log.Fatalf("insert: %v", err)
	}
	fmt.Printf("Inserted: %+v\n", inserted)

	// 3) Read it back by ID
	got, err := getTradeByID(ctx, pool, inserted.ID)
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	fmt.Printf("Fetched:  %+v\n", got)
}

func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS trades (
			id          BIGSERIAL PRIMARY KEY,
			market      TEXT NOT NULL,
			side        TEXT NOT NULL CHECK (side IN ('YES','NO')),
			price       DOUBLE PRECISION NOT NULL CHECK (price >= 0 AND price <= 1),
			quantity    INTEGER NOT NULL CHECK (quantity > 0),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		);

		CREATE INDEX IF NOT EXISTS idx_trades_market_created_at
		ON trades (market, created_at DESC);
	`)
	return err
}

func insertTrade(ctx context.Context, pool *pgxpool.Pool, t Trade) (Trade, error) {
	// Return the inserted row (including id + created_at) in one round-trip.
	err := pool.QueryRow(ctx, `
		INSERT INTO trades (market, side, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, t.Market, t.Side, t.Price, t.Quantity).Scan(&t.ID, &t.CreatedAt)

	return t, err
}

func getTradeByID(ctx context.Context, pool *pgxpool.Pool, id int64) (Trade, error) {
	var t Trade
	err := pool.QueryRow(ctx, `
		SELECT id, market, side, price, quantity, created_at
		FROM trades
		WHERE id = $1
	`, id).Scan(&t.ID, &t.Market, &t.Side, &t.Price, &t.Quantity, &t.CreatedAt)

	return t, err
}
