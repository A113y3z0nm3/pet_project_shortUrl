package client

import (
	"context"
	"fmt"
	"short_url/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxClient создает клиент-подключение к базе данных пользователей PostgreSQL
func NewPgxClient(ctx context.Context, config *models.ConfigDB) (*pgxpool.Pool, error) {

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.User, config.Password, config.Host, config.Port, config.Database)
	pool, err := pgxpool.New(ctx, databaseUrl)

	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
