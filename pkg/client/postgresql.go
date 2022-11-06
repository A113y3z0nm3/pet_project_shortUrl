package client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxClient создает клиент-подключение к базе данных пользователей PostgreSQL
func NewPgxClient(ctx context.Context, host, port, database, username, password string) (*pgxpool.Pool, error) {

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	pool, err := pgxpool.New(ctx, databaseUrl)

	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
