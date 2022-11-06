package client

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
)

// NewRedisClient создает клиент-подключение к базе данных ссылок Redis
func NewRedisClient(ctx context.Context, path, host, port, username, password string, database int) (*redis.Client, error) {
	var databaseUrl string

	if path == ""{
		// tcp conn
		databaseUrl = fmt.Sprintf("redis://%s:%s@%s:%s/%v", username, password, host, port, database)
	} else {
		// unix conn
		databaseUrl = fmt.Sprintf("unix://%s:%s@%s?db=%v", username, password, path, database)
	}

	// Парсит настройки из URL базы данных
	opts, err := redis.ParseURL(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %e", err)
	}
	
	// Создает клиента с настройками
	rdb := redis.NewClient(opts)

	// "Пингует" по БД (проверка)
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis: %e", err)
	}
	
	return rdb, nil
}
