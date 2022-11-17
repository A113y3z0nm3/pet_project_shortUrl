package handlers

import (
	"bytes"
	"context"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// linkService Интерфейс к сервису, осуществляющему управление пользователя ссылками
type linkService interface {
	FindLink(ctx context.Context, link string) (models.LinkDataDTO, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDTO, error)
	DeleteLink(ctx context.Context, username, link string) error
	CreateLink(ctx context.Context, fullUrl, custom string, exp int, user models.JWTUserInfo) (models.LinkDataDTO, error)
	CreateQR(ctx context.Context, url, link string) (*bytes.Buffer, error)
}

// LinkHandlerConfig Конфигурация для LinkHandler
type LinkHandlerConfig struct {
	Router			*gin.Engine
	LinkService		linkService
	Middleware		*middlewares.Middlewares
	Logger			*myLog.Log
}

// LinkHandler Для логирования и регистрации хендлеров
type LinkHandler struct {
	linkService linkService
	middleware		*middlewares.Middlewares
	logger			*myLog.Log
}

// getLinkFromParam Получает короткую ссылку из path
func getLinkFromParam(ctx *gin.Context) string {
	link := ctx.Param("link")

	return link
}

// RegisterLinkHandler Фабрика для LinkHandler
func RegisterLinkHandler(c *LinkHandlerConfig) {
	linkHandler := LinkHandler{
		linkService:	c.LinkService,
		middleware:		c.Middleware,
		logger:			c.Logger,
	}

	g := c.Router.Group("v1")
	g.POST("/newlink", c.Middleware.Recorder, c.Middleware.AuthUser, linkHandler.CreateLink)
	g.DELETE("/links/:link", c.Middleware.Recorder, c.Middleware.AuthUser, linkHandler.DeleteLink)
	g.GET("/links", c.Middleware.Recorder, c.Middleware.AuthUser, linkHandler.GetAllLinks)
	g.GET("/:link", c.Middleware.Recorder, linkHandler.LinkRedirect)
	g.GET("/links/qr/:link", c.Middleware.Recorder, c.Middleware.AuthUser, linkHandler.CreateCode)
	g.GET("/links/:link", c.Middleware.Recorder, c.Middleware.AuthUser, linkHandler.GetLink)
}
