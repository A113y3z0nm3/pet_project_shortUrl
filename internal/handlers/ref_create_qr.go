package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// CreateCode Создает QR-код из ссылки
func (h *LinkHandler) CreateCode(ctx *gin.Context) {

	// Получаем информацию о пользователе
	_, err := h.GetUserInfo(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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
