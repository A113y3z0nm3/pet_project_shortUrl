package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// LinkData Структура данных для одной ссылки
type LinkData struct {
	Short   string
	Full    string
	ExpTime string
}

// getAllLinksResponse Ответ на запрос
type getAllLinksResponse struct {
	Data []LinkData
}

// GetAllLinks Отдает все ссылки пользователя
func (h *ManageHandler) GetAllLinks(ctx *gin.Context) {

	// Получаем информацию о пользователе
	user := h.GetUserInfo(ctx)
	if user.Err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": user.Err.Error(),
		})

		return
	}

	// Получаем все ссылки пользователя
	data, err := h.manageService.GetAllLinks(ctx, user.Username)
	if err != nil {
		if err != redis.Nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})

			return
		} else {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "links not found",
			})

			return
		}
	}

	// Маппим данные в ответ
	resp := getAllLinksResponse{
		Data: make([]LinkData, len(data)),
	}
	for k, l := range data {
		linkData := LinkData{
			Short:   l.Link,
			Full:    l.FullURL,
			ExpTime: fmt.Sprint(l.ExpTime),
		}
		resp.Data[k] = linkData
	}

	ctx.JSON(http.StatusOK, resp)

	return
}
