package client

import (
	"context"
	"errors"
	"fmt"
	"short_url/internal/models"

	"github.com/go-redis/redis/v9"
)

// NewRedisClient создает клиент-подключение к базе данных ссылок Redis
func NewRedisClient(ctx context.Context, config *models.ConfigRedis) (*redis.Client, error) {

	// Собираем адрес redis
	var databaseUrl string
	if (config.Host != "" && config.Port != "") {
		// tcp conn
		databaseUrl = fmt.Sprintf("redis://%s:%s@%s:%s/%v", config.User, config.Password, config.Host, config.Port, config.Database)
	} else if config.Path != "" {
		// unix conn
		databaseUrl = fmt.Sprintf("unix://%s:%s@%s?db=%v", config.User, config.Password, config.Path, config.Database)
	} else {
		// conn not established
		return nil, errors.New("failed to build redis URL")
	}

	// Парсим настройки из адреса redis
	opts, err := redis.ParseURL(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %e", err)
	}
	
	// Создаем клиента с настройками
	rdb := redis.NewClient(opts)

	// "Пингуем" по БД (проверка)
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis: %e", err)
	}
	
	return rdb, nil
}
