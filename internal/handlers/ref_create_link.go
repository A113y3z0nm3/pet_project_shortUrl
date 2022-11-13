package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// createLinkRequest Структура запроса
type createLinkRequest struct {
	Full    string
	ExpTime int
	Custom  string
}

// createLinkResponse Структура ответа
type createLinkResponse struct {
	Link    string
	Full    string
	ExpTime string
}

// CreateLink Создает ссылку
func (h *LinkHandler) CreateLink(ctx *gin.Context) {

	var req createLinkRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Получаем информацию о пользователе
	user, err := h.middleware.GetUserInfo(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Создаем ссылку
	data, err := h.manageService.CreateLink(ctx, req.Full, req.Custom, req.ExpTime, user)
	if err != nil {
		if err.Error() != "limit exceeded" {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})

			return
		} else {
			ctx.JSON(http.StatusNoContent, gin.H{
				"error": "limit exceeded: maximum links",
			})

			return
		}
	}

	// Маппим данные в ответ
	resp := createLinkResponse{
		Link:		data.Link,
		Full:		data.FullURL,
		ExpTime:	fmt.Sprint(data.ExpTime),
	}

	ctx.JSON(http.StatusOK, resp)

	return
}
