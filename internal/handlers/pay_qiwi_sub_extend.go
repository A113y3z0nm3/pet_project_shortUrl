package handlers

import (
	"net/http"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// QiwiSubExtend Создает счет в QIWI P2P для продления подписки
func (h *PayHandler) QiwiSubExtend(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "QiwiSubExtendHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("QiwiSubExtendHandler() started")
	defer l.Debug("QiwiSubExtendHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	// Получаем информацию о пользователе
	user, err := GetUserInfo(ctx)
	if err != nil {
		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "GET", MetricQiwiSubExt)

		return
	}

	// Если пользователь не подписчик, закрываем ручку и отдаем ответ
	if user.Subscribe != models.Sub {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "user does not have a subscribe",
		})

		Bridge(ctx, http.StatusForbidden, "GET", MetricQiwiSubExt)

		return
	}

	// Получаем желаемый срок подписки из path
	sub, err := h.getSubFromParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid subscribe params",
		})

		Bridge(ctx, http.StatusBadRequest, "GET", MetricQiwiSubExt)

		return
	}

	// Создаем счет на оплату и получаем ссылку на него
	result, err := h.qiwiService.BillRequest(ctx, sub, user.Username)
	if err != nil {
		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "GET", MetricQiwiSubExt)

		return
	}

	// Отдаем ссылку на счет клиенту
	ctx.JSON(http.StatusOK, gin.H{
		"payUrl": result,
	})

	Bridge(ctx, http.StatusOK, "GET", MetricQiwiSubExt)

	return
}
