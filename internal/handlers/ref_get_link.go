package handlers

import (
	"fmt"
	"net/http"
	"short_url/internal/handlers/middlewares"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// getLinkResponse Ответ на запрос
type getLinkResponse struct {
	Short   string	`json:"short"`
	Full    string	`json:"full"`
	ExpTime string	`json:"time"`
}

// GetLink Отдает ссылку и информацию о ней
func (h *LinkHandler) GetLink(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "GetLinkHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("GetLinkHandler() started")
	defer l.Debug("GetLinkHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	// Получаем короткую ссылку из path
	link := getLinkFromParam(ctx)

	// Ищем данные связанные с этой ссылкой, проверяем валидность
	data, err := h.linkService.FindLink(ctx, link)
	if err != nil {
		if err != redis.Nil {
			InternalErrResp(ctx, l, err)

			Bridge(ctx, http.StatusInternalServerError, "GET", MetricGetLink)

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})

			Bridge(ctx, http.StatusNotFound, "GET", MetricGetLink)

			return
		}
	}

	// Маппим данные в ответ
	resp := getLinkResponse{
		Short:   data.Link,
		Full:    data.FullURL,
		ExpTime: fmt.Sprint(data.ExpTime),
	}

	ctx.JSON(http.StatusOK, resp)

	Bridge(ctx, http.StatusOK, "GET", MetricGetLink)

	return
}
