package handlers

import (
	"context"
	"errors"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// qiwiService Интерфейс к сервису оплаты подписок через QIWI
type qiwiService interface {
	NotifyFromQiwi(ctx context.Context, status, bill string) error
	BillRequest(ctx context.Context, amo float64, username string) (string, error)
}

// PayHandlerConfig Конфигурация к PayHandler
type PayHandlerConfig struct {
	Router		*gin.Engine
	Logger		*myLog.Log
	QiwiService	qiwiService
	Middleware	*middlewares.Middlewares
	Prices		models.SubPrice
}

// PayHandler Для регистрации "ручек"
type PayHandler struct {
	logger		*myLog.Log
	qiwiService	qiwiService
	middleware	*middlewares.Middlewares
	prices		models.SubPrice
}

// getSubFromParam Получает вариант подписки из path и вычисляет на какую сумму выставить счет
func getSubFromParam(ctx *gin.Context, prices models.SubPrice) (float64, error) {
	sub := ctx.Param("subTime")

	var amo float64
	var err error

	switch sub {
	case "weekly":
		amo = prices.Weekly
	case "monthly":
		amo = prices.Monthly
	case "yearly":
		amo = prices.Yearly
	default:
		err = errors.New("param error")
	}
	return amo, err
}

// RegisterPayHandler Фабрика для PayHandler
func RegisterPayHandler(c *PayHandlerConfig) {
	payHandler := &PayHandler{
		logger:			c.Logger,
		qiwiService:	c.QiwiService,
		middleware:		c.Middleware,
		prices:			c.Prices,
	}

	g := c.Router.Group("v1") // Версия API
	g.GET("/qiwi/:subTime", c.Middleware.AuthUser, payHandler.QiwiSub)
	g.POST("/qiwistatus", payHandler.QiwiNotify, c.Middleware.QiwiAuthorization)
	g.GET("/qiwi/extend/:subTime")
}