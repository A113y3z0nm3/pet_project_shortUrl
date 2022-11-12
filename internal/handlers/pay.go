package handlers

import (
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

//
type payService interface {
	SaveBillId(billId string, info models.SubInfo)
	CalculateSub(amount float64, username string) (models.SubInfo, error)
}

//
type PayHandlerConfig struct {
	Router		*gin.Engine
	Logger		*myLog.Log
	PayService	payService
	Middleware	*middlewares.Middlewares
	Prices		models.SubPrice
}

//
type PayHandler struct {
	logger		*myLog.Log
	payService	payService
	middleware	*middlewares.Middlewares
	prices		models.SubPrice
}

//
func RegisterPayHandler(c *PayHandlerConfig) {
	payHandler := &PayHandler{
		logger: 	c.Logger,
		payService: c.PayService,
		middleware: c.Middleware,
		prices:		c.Prices,
	}

	g := c.Router.Group("v1") // Версия API
	g.GET("/qiwi", c.Middleware.AuthUser, payHandler.Qiwi)
	g.GET("/qiwi/:subTime", c.Middleware.AuthUser, payHandler.QiwiSub)
	g.POST("/qiwi/status", payHandler.QiwiNotify, c.Middleware.QiwiAuthorization)
}