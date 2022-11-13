package middlewares

import (
	"context"
	"errors"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	UserInfo	= "user_info"	// Ключ для контекста (информация о пользователе)
)

// GetUserInfo Возвращает информацию о пользователе и его подписке из контекста
func (m *Middlewares) GetUserInfo(ctx *gin.Context) (models.JWTUserInfo, error) {

	// Получаем данные из контекста
	info, ok := ctx.Get(UserInfo)
	if !ok {
		m.logger.Error("failed to get user info from ctx")

		return models.JWTUserInfo{}, errors.New("failed to get user info")
	}

	// Преобразуем их в структуру для ответа
	user, ok := info.(models.JWTUserInfo)
	if !ok {
		m.logger.Error("failed to get JWT user info")

		return models.JWTUserInfo{}, errors.New("failed to JWT user info")
	}

	return user, nil
}

// tokenService Интерфейс к сервису управления токенами
type tokenService interface {
	ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error)
}

// Middlewares класс для работы с middlewares
type Middlewares struct {
	basicPassword	string
	key				string
	tokenService 	tokenService
	logger          *myLog.Log
}

// NewMiddlewares конструктор для Middlewares
func NewMiddlewares(log *myLog.Log, service tokenService, key, basicPassword string) *Middlewares {

	return &Middlewares{
		basicPassword:	basicPassword,
		key:			key,
		tokenService:	service,
		logger:			log,
	}
}
