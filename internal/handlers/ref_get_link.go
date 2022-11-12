package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// getLinkResponse Ответ на запрос
type getLinkResponse struct {
	Short   string
	Full    string
	ExpTime string
}

// GetLink Отдает ссылку и информацию о ней
func (h *LinkHandler) GetLink(ctx *gin.Context) {

	// Получаем короткую ссылку из path
	link := getLinkFromParam(ctx)

	// Получаем информацию о пользователе
	_, err := h.GetUserInfo(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Ищем данные связанные с этой ссылкой, проверяем валидность
	data, err := h.manageService.FindLink(ctx, link)
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

	// Маппим данные в ответ
	resp := getLinkResponse{
		Short:   data.Link,
		Full:    data.FullURL,
		ExpTime: fmt.Sprint(data.ExpTime),
	}

	ctx.JSON(http.StatusOK, resp)

	return
}
