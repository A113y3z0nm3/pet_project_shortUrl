package handlers

import "github.com/gin-gonic/gin"

// qiwiSendResponse Структура ответа
type qiwiSendResponse struct {
	Weekly	float64
	Monthly	float64
	Yearly	float64
}

// QiwiSend Отправляет информацию о стоимости подписки
func (h *PayHandler) Qiwi(ctx *gin.Context) {
	
}