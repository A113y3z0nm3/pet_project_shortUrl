package middlewares

import (
	"context"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"
)

const (
	UserInfo	= "user_info"	// Ключ для контекста (информация о пользователе)
)

// tokenService Интерфейс к сервису управления токенами
type tokenService interface {
	ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error)
}

// Middlewares класс для работы с middlewares
type Middlewares struct {
	basicPassword string
	tokenService  tokenService
	log           *myLog.Log
}

// NewMiddlewares конструктор для Middlewares
func NewMiddlewares(log *myLog.Log, service tokenService, basicPassword string) *Middlewares {

	return &Middlewares{
		basicPassword: basicPassword,
		tokenService:  service,
		log:           log,
	}
}
