package handlers

import (
	"net/http"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LinkRedirect Выполняет переадресацию на источник при переходе на короткую ссылку
func (h *LinkHandler) LinkRedirect(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "LinkRedirectHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("LinkRedirectHandler() started")
	defer l.Debug("LinkRedirectHandler() done")

	// Получаем короткую ссылку
	link := getLinkFromParam(ctx)

	// Ищем данные связанные с этой ссылкой, проверяем валидность
	data, err := h.linkService.FindLink(ctx, link)
	if err != nil {
		if err.Error() != "not found" {
			InternalErrResp(ctx, l, err)

			Bridge(ctx, http.StatusInternalServerError, "GET", MetricRedirectLink)

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})

			Bridge(ctx, http.StatusNotFound, "GET", MetricRedirectLink)

			return
		}
	}

	// Переадресовываем пользователя на источник
	ctx.Redirect(http.StatusOK, data.FullURL)

	Bridge(ctx, http.StatusOK, "GET", MetricRedirectLink)

	return
}
