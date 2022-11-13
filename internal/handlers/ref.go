package handlers

import (
	"context"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"

	"github.com/boombuler/barcode"
	"github.com/gin-gonic/gin"
)

// linkService Интерфейс к сервису, осуществляющему управление пользователя ссылками
type linkService interface {
	FindLink(ctx context.Context, link string) (models.LinkDataDTO, error)
	GetAllLinks(ctx context.Context, username string) ([]models.LinkDataDTO, error)
	DeleteLink(ctx context.Context, username, link string) error
	CreateLink(ctx context.Context, fullUrl, custom string, exp int, user models.JWTUserInfo) (models.LinkDataDTO, error)
	CreateQR(ctx context.Context, url, link string) (barcode.Barcode, error)
}

// LinkHandlerConfig Конфигурация для LinkHandler
type LinkHandlerConfig struct {
	Router        *gin.Engine
	ManageService linkService
	Middlware		*middlewares.Middlewares
	Logger			*myLog.Log
}

// LinkHandler Для логирования и регистрации хендлеров
type LinkHandler struct {
	manageService linkService
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
		manageService: c.ManageService,
		middleware:		c.Middlware,
		logger:			c.Logger,
	}

	g := c.Router.Group("v1")
	g.POST("/newlink",c.Middlware.AuthUser, linkHandler.CreateLink)
	g.DELETE("/links/:link",c.Middlware.AuthUser, linkHandler.DeleteLink)
	g.GET("/links",c.Middlware.AuthUser, linkHandler.GetAllLinks)
	g.GET("/:link", linkHandler.LinkRedirect)
	g.GET("/links/qr/:link",c.Middlware.AuthUser, linkHandler.CreateCode)
	g.GET("/links/:link",c.Middlware.AuthUser, linkHandler.GetLink)
}
