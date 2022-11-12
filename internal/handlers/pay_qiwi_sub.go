package handlers

import (
	"errors"
	"short_url/internal/models"

	"github.com/gin-gonic/gin"
)

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

//
type qiwiSubResponse struct {
	url		string
}

//
func (h *PayHandler) QiwiSub(ctx *gin.Context) {
	//
	sub, err := getSubFromParam(ctx, h.prices)
	if err != nil {
		//
	}



	return
}