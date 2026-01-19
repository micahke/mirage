package clients

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresClient provides an interface for PostgreSQL database operations.
// Methods return pgx types directly to allow scanning into any struct (including protobufs).
type PostgresClient interface {
	// QueryRow executes a query that returns at most one row.
	// Use .Scan() on the returned Row to read values.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Query executes a query that returns multiple rows.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE).
	// Returns the command tag with rows affected.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

	// BeginTx starts a transaction.
	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Ping verifies the connection is alive.
	Ping(ctx context.Context) error

	// Close closes all connections in the pool.
	Close()
}

type postgresClient struct {
	pool *pgxpool.Pool
}

// NewPostgresClient creates a new PostgreSQL client with connection pooling.
// The dsn should be a PostgreSQL connection string, e.g.:
// "postgres://user:password@localhost:5432/dbname?sslmode=disable"
func NewPostgresClient(ctx context.Context, dsn string) (PostgresClient, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	fmt.Println("Connected to PostgreSQL")
	return &postgresClient{pool: pool}, nil
}

func (p *postgresClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *postgresClient) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *postgresClient) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}

func (p *postgresClient) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

func (p *postgresClient) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *postgresClient) Close() {
	p.pool.Close()
}

// IsNoRows checks if the error is pgx.ErrNoRows (no rows returned from query).
func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
