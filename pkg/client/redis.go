package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
)

// NewRedisClient создает клиент-подключение к базе данных ссылок Redis
func NewRedisClient(ctx context.Context, path, host, port, username, password string, database int) (*redis.Client, error) {

	// Собираем адрес redis
	var databaseUrl string
	if (host != "" && port != "") {
		// tcp conn
		databaseUrl = fmt.Sprintf("redis://%s:%s@%s:%s/%v", username, password, host, port, database)
	} else if path != "" {
		// unix conn
		databaseUrl = fmt.Sprintf("unix://%s:%s@%s?db=%v", username, password, path, database)
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
