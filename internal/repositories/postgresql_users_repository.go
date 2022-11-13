package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"short_url/internal/models"
)

// PostgresqlUserRepositoryConfig конфигурация для PostgresqlUserRepository
type PostgresqlUserRepositoryConfig struct {
	Table string
	DB    *pgxpool.Pool
}

// PostgresqlUserRepository - слой для управления запросами к Postgresql
type PostgresqlUserRepository struct {
	table string
	db    *pgxpool.Pool
}

// NewPostgresqlUserRepository конструктор для PostgresqlUserRepository
func NewPostgresqlUserRepository(c *PostgresqlUserRepositoryConfig) *PostgresqlUserRepository {
	return &PostgresqlUserRepository{
		table: c.Table,
		db:    c.DB,
	}
}

// FindByUsername ищет пользователя по его имени
func (r *PostgresqlUserRepository) FindByUsername(ctx context.Context, username string) (models.UserDB, error) {
	var user models.UserDB

	query := fmt.Sprintf("SELECT * FROM %s WHERE username = $1", r.table)

	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Password)

	if err != nil {
		return user, err
	}

	return user, nil
}

// CreateUser создает пользователя в базе данных
func (r *PostgresqlUserRepository) CreateUser(ctx context.Context, user models.UserDB) error {
	var id int64

	query := fmt.Sprintf("INSERT INTO %s (username, first_name, last_name, password) VALUES ($1, $2, $3, $4) RETURNING user_id", r.table)

	err := r.db.QueryRow(ctx, query, user.Username, user.FirstName, user.LastName, user.Password).Scan(&id)

	if err != nil {
		return err
	}

	return nil
}
