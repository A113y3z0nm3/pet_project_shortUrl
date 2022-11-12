package repositories

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
)

// RedisLinkRepositoryConfig Конфигурация для RedisLinkRepository
type RedisSubRepositoryConfig struct {
	DB		*redis.Client
}

// RedisLinkRepository Слой для управления запросами к хранилищу ссылок
type RedisSubRepository struct {
	db		*redis.Client
	pipe 	redis.Pipeliner
}

func NewRedisSubRepository(c *RedisSubRepositoryConfig) *RedisSubRepository {
	redisRepo := &RedisSubRepository{
		db:		c.DB,
	}

	// Инициализируем пайплайн (выполняет несколько команд за одну запись)
	redisRepo.pipe = redisRepo.db.Pipeline()

	return redisRepo
}

func (r *RedisSubRepository) AddSubRedis(ctx context.Context, username string, exp time.Duration) error {
	// Загружаем информацию о подписке в базу
	_, err := r.pipe.Set(ctx, username, 1, exp).Result()
	if err != nil {
		return err
	}

	return nil
}