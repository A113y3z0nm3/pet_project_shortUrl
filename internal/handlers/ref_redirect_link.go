package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LinkRedirect Выполняет переадресацию на источник при переходе на короткую ссылку
func (h *LinkHandler) LinkRedirect(ctx *gin.Context) {

	// Получаем короткую ссылку
	link := getLinkFromParam(ctx)

	// Ищем данные связанные с этой ссылкой, проверяем валидность
	data, err := h.manageService.FindLink(ctx, link)
	if err != nil {
		if err.Error() != "not found" {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})

			return
		}
	}

	// Переадресовываем пользователя на источник
	ctx.Redirect(http.StatusOK, data.FullURL)

	return
}