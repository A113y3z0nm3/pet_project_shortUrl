package handlers

import (
	"fmt"
	"net/http"
	"short_url/internal/handlers/middlewares"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// LinkData Структура данных для одной ссылки
type LinkData struct {
	Short   string	`json:"short"`
	Full    string	`json:"full"`
	ExpTime string	`json:"time"`
}

// getAllLinksResponse Ответ на запрос
type getAllLinksResponse struct {
	Data []LinkData	`json:"data"`
}

// GetAllLinks Отдает все ссылки пользователя
func (h *LinkHandler) GetAllLinks(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "GetAllLinksHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("GetAllLinksHandler() started")
	defer l.Debug("GetAllLinksHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	// Получаем информацию о пользователе
	user, err := GetUserInfo(ctx)
	if err != nil {
		InternalErrResp(ctx, l ,err)

		Bridge(ctx, http.StatusInternalServerError, "GET", MetricGetAllLinks)

		return
	}

	// Получаем все ссылки пользователя
	data, err := h.linkService.GetAllLinks(ctx, user.Username)
	if err != nil {
		if err != redis.Nil {
			InternalErrResp(ctx, l, err)

			Bridge(ctx, http.StatusInternalServerError, "GET", MetricGetAllLinks)

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "links not found",
			})

			Bridge(ctx, http.StatusNotFound, "GET", MetricGetAllLinks)

			return
		}
	}

	// Маппим данные в ответ
	resp := getAllLinksResponse{
		Data: make([]LinkData, len(data)),
	}
	for k, l := range data {
		linkData := LinkData{
			Short:   l.Link,
			Full:    l.FullURL,
			ExpTime: fmt.Sprint(l.ExpTime),
		}
		resp.Data[k] = linkData
	}

	ctx.JSON(http.StatusOK, resp)

	Bridge(ctx, http.StatusOK, "GET", MetricGetAllLinks)

	return
}
