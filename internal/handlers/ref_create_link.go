package handlers

import (
	"fmt"
	"net/http"
	"short_url/internal/handlers/middlewares"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// createLinkRequest Структура запроса
type createLinkRequest struct {
	Full    string	`json:"full" binding:"required"`
	ExpTime int		`json:"time" binding:"required"`
	Custom  string	`json:"custom"`
}

// createLinkResponse Структура ответа
type createLinkResponse struct {
	Link    string	`json:"link"`
	Full    string	`json:"full"`
	ExpTime string	`json:"time"`
}

// CreateLink Создает короткую ссылку
func (h *LinkHandler) CreateLink(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "CreateLinkHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("CreateLinkHandler() started")
	defer l.Debug("CreateLinkHandler() done")

	// Если был получен сигнал пропускаем ручку для обработки метрик
	_, ok := ctx.Get(middlewares.Skip) 
	if ok {
		ctx.Next()
	}

	var req createLinkRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, l, &req, "POST", MetricCreateLink); !ok {
		return
	}

	// Получаем информацию о пользователе
	user, err := GetUserInfo(ctx)
	if err != nil {
		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "POST", MetricCreateLink)

		return
	}

	// Создаем ссылку
	data, err := h.linkService.CreateLink(ctx, req.Full, req.Custom, req.ExpTime, user)
	if err != nil {
		if err.Error() == "need subscribe" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "need subscribe",
			})

			Bridge(ctx, http.StatusForbidden, "POST", MetricCreateLink)

			return
		}

		if err.Error() != "limit exceeded" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "limit exceeded: maximum links",
			})
	
			Bridge(ctx, http.StatusForbidden, "POST", MetricCreateLink)
	
			return
		} 

		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "POST", MetricCreateLink)

		return
	}

	// Маппим данные в ответ
	resp := createLinkResponse{
		Link:		data.Link,
		Full:		data.FullURL,
		ExpTime:	fmt.Sprint(data.ExpTime),
	}

	ctx.JSON(http.StatusOK, resp)

	Bridge(ctx, http.StatusOK, "POST", MetricCreateLink)

	return
}
