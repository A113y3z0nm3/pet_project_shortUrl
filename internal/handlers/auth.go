package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"
)

// Интерфейс для сервиса, который управляет регистрацией и входом пользователей
type authService interface {
	SignInUserByName(ctx context.Context, user models.SignInUserDTO) (models.SignInUserDTO, error)
	SignUpUser(ctx context.Context, user models.SignUpUserDTO) error
}

// Интерфейс для сервиса, который управляет токенами доступа
type tokenService interface {
	ValidateToken(ctx context.Context, token string) (models.JWTUserInfo, error)
	CreateToken(ctx context.Context, dto models.CreateTokenDTO) (string, error)
}

// AuthHandlerConfig конфигурация для AuthHandler
type AuthHandlerConfig struct {
	Router       *gin.Engine
	AuthService  authService
	TokenService tokenService
	Middleware   *middlewares.Middlewares
	Logger       *myLog.Log
}

// AuthHandler для регистрации "ручек"
type AuthHandler struct {
	authService  authService
	tokenService tokenService
	middleware	 *middlewares.Middlewares
	logger       *myLog.Log
}

// RegisterAuthHandler фабрика для AuthHandler
func RegisterAuthHandler(c *AuthHandlerConfig) {
	authHandler := AuthHandler{
		authService:  c.AuthService,
		tokenService: c.TokenService,
		middleware:	  c.Middleware,
		logger:       c.Logger,
	}

	g := c.Router.Group("v1") // Версия API
	g.POST("/signin", authHandler.SignIn)
	g.POST("/signup", authHandler.SignUp)
}
