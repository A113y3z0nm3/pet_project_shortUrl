package handlers

import (
	"net/http"
	"short_url/internal/handlers/middlewares"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// DeleteLink Удаляет короткую ссылку
func (h *LinkHandler) DeleteLink(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "DeleteLinkHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("DeleteLinkHandler() started")
	defer l.Debug("DeleteLinkHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	// Получаем информацию о пользователе
	user, err := GetUserInfo(ctx)
	if err != nil {
		InternalErrResp(ctx, l ,err)

		Bridge(ctx, http.StatusInternalServerError, "DELETE", MetricDeleteLink)

		return
	}

	// Получаем короткую ссылку
	link := getLinkFromParam(ctx)

	// Удаляем ссылку
	err = h.linkService.DeleteLink(ctx, user.Username, link)
	if err != nil {
		if err != redis.Nil {
			InternalErrResp(ctx, l ,err)

			Bridge(ctx, http.StatusInternalServerError, "DELETE", MetricDeleteLink)

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})

			Bridge(ctx, http.StatusNotFound, "DELETE", MetricDeleteLink)

			return
		}
	}

	ctx.JSON(http.StatusOK, "OK")

	Bridge(ctx, http.StatusOK, "DELETE", MetricDeleteLink)

	return
}
