package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// CreateCode Создает QR-код из ссылки
func (h *ManageHandler) CreateCode(ctx *gin.Context) {

	// Получаем информацию о пользователе
	user := h.GetUserInfo(ctx)
	if user.Err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": user.Err.Error(),
		})

		return
	}

	// Получаем короткую ссылку
	link := getLinkFromParam(ctx)

	// Конструируем веб-ссылку
	url := ctx.Request.URL.Scheme + "//" + ctx.Request.Host + "/" + link

	// Создаем QR-код
	qrCode, err := h.manageService.CreateQR(ctx, url, link)
	if err != nil {
		if err != redis.Nil {
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

	ctx.JSON(http.StatusOK, qrCode)

	return
}
