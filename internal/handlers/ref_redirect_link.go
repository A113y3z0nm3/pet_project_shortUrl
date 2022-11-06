package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getLinkFromParam Получает короткую ссылку из path
func getLinkFromParam(ctx *gin.Context) string {
	link := ctx.Param("link")

	return link
}

// LinkRedirect Выполняет переадресацию на источник при переходе на короткую ссылку
func (h *ManageHandler) LinkRedirect(ctx *gin.Context) {

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