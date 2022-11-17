package handlers

import (
	"net/http"
	"short_url/internal/handlers/middlewares"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// CreateCode Создает QR-код из ссылки
func (h *LinkHandler) CreateCode(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "CreateCodeHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("CreateCodeHandler() started")
	defer l.Debug("CreateCodeHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	// Получаем короткую ссылку
	link := getLinkFromParam(ctx)

	// Конструируем веб-ссылку
	url := ctx.Request.URL.Scheme + "//" + ctx.Request.Host + "/" + link

	// Создаем QR-код
	byteQrCode, err := h.linkService.CreateQR(ctx, url, link)
	if err != nil {
		if err != redis.Nil {
			InternalErrResp(ctx, l, err)

			Bridge(ctx, http.StatusInternalServerError, "GET", MetricCreateQR)

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})

			Bridge(ctx, http.StatusNotFound, "GET", MetricCreateQR)

			return
		}
	}

	result := byteQrCode.Bytes()

	ctx.JSON(http.StatusOK, result)

	Bridge(ctx, http.StatusOK, "GET", MetricCreateQR)

	return
}
