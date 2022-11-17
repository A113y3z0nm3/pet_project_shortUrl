package middlewares

import (
	"context"
	"short_url/internal/models"
	log "short_url/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
)

// Ключи для контекста
const (
	UserInfo	= "user_info"		//  Информация о пользователе

	Method		= "method"			// Метод запроса
	Code		= "code"			// Код ответа
	Handler		= "handler"			// Обработчик

	MidAuth		= "middleware_auth"	// Проверка авторизации
	Skip		= "skipped"			// Ключ пропуска обработчика (вместо ctx.Abort())
)

// tokenService Интерфейс к сервису управления токенами
type tokenService interface {
	ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error)
}

// Middlewares класс для работы с middlewares
type Middlewares struct {
	tokenService 	tokenService
	logger          *log.Log
	Counter			*prometheus.CounterVec
}

// NewMiddlewares конструктор для Middlewares
func NewMiddlewares(log *log.Log, service tokenService) *Middlewares {
	// Создаем метрику
	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_total",
			Help: "Кол-во запросов ко всем эндпоинтам",
		},
		[]string{"handler", "method", "code"},
	)

	return &Middlewares{
		tokenService:	service,
		logger:			log,
		Counter:		requestTotal,
	}
}
