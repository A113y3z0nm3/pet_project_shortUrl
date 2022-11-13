package repositories

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
)

// RedisSubRepositoryConfig Конфигурация для RedisLinkRepository
type RedisSubRepositoryConfig struct {
	Db		*redis.Client
	Pipe	redis.Pipeliner
}

// RedisSubRepository Слой для управления запросами к хранилищу ссылок
type RedisSubRepository struct {
	db		*redis.Client
	pipe 	redis.Pipeliner
}

// NewRedisSubRepository Конструктор для RedisSubRepository
func NewRedisSubRepository(c *RedisSubRepositoryConfig) *RedisSubRepository {
	return &RedisSubRepository{
		db:		c.Db,
		pipe:	c.Pipe,
	}
}

// FindByUsername Находит подписку и ее срок по имени пользователя
func (r *RedisSubRepository) FindByUsername(ctx context.Context, username string) (time.Duration, bool) {
	// Ищем в базе информацию о подписке
	_, err := r.db.Get(ctx, username).Result()
	if err != nil {
		return -1, false
	}

	exp, err := r.db.TTL(ctx, username).Result()
	if err != nil {
		return -1, false
	}

	return exp, true
}

// AddSubRedis Добавляет пользователю подписку на ограниченное время
func (r *RedisSubRepository) AddSubRedis(ctx context.Context, username string, exp time.Duration) error {
	// Загружаем информацию о подписке в базу
	_, err := r.db.Set(ctx, username, 1, exp).Result()
	if err != nil {
		return err
	}

	return nil
}