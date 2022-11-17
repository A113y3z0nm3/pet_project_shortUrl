package models

import (
	"crypto/rsa"
)

// Config основная конфигурация сервиса
type Config struct {
	App		*ConfigApp
	Price	*ConfigPrice
	Log		*ConfigLog
	HTTP	*ConfigHTTP
	DB		*ConfigDB
	RDB		*ConfigRedis
	JWT		*ConfigJWT
}

// ConfigHTTP конфигурация для HTTP
type ConfigHTTP struct {
	Host       string `env:"HTTP_HOST"`         //  HTTP хост
	Port       string `env:"HTTP_PORT"`         //  HTTP порт
	MetricPort string `env:"HTTP_METRICS_PORT"` //  Prometheus порт
}

// ConfigDB конфигурация для подключения к базе данных
type ConfigDB struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Database string `env:"DB_DATABASE"`
}

// ConfigRedis конфигурация для подключения к Redis
type ConfigRedis struct {
	Path		string	`env:"REDIS_PATH"`
	Host		string	`env:"REDIS_HOST"`
	Port		string	`env:"REDIS_PORT"`
	User		string	`env:"REDIS_USER"`
	Password	string	`env:"REDIS_PASSWORD"`
	Database	int		`env:"REDIS_DATABASE"`
}

// ConfigApp конфигурация для внутренних модулей приложения
type ConfigApp struct {
	SecretKey     string `env:"SECRET_KEY"`
}

// ConfigPrice Стоимость подписок
type ConfigPrice struct {
	// Недельная
	Weekly float64	`env:"WEEK_PRICE"`
	// Месячная
	Monthly float64	`env:"MONTH_PRICE"`
	// Годовая
	Yearly float64	`env:"YEAR_PRICE"`
}

// ConfigLog конфигурация для логгера
type ConfigLog struct {
	Mode	string	`env:"LOG_MODE"`
	Level	string	`env:"LOG_LEVEL"`
	Output	string	`env:"LOG_OUTPUT"`
}

// ConfigJWT конфигурация для создания токенов авторизации
type ConfigJWT struct {
	PublicKey             *rsa.PublicKey
	PrivateKey            *rsa.PrivateKey
	AccessTokenExpiration int64 `env:"JWT_ACCESS_TOKEN_EXPIRATION"`
}
