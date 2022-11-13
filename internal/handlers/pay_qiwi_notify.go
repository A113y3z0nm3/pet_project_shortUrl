package handlers

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
)

// customFields Дополнительные данные счета (нет)
type customFields struct {}

// customer Данные о пользователе, на которого был выставлен счет
type customer struct {
	Phone	string
	Email	string
	Account	string
}

// status Данные о статусе счета
type status struct {
	Value			string
	ChangedDatetime	string
}

// amount Данные о сумме счета
type amount struct {
	Value		float64
	Currency	string
}

// bill Данные о счете
type bill struct {
	SiteId				string
	BillID				string
	Amount				amount
	Status				status
	Customer			customer
	CustomFields		customFields
	CreationDateTime	string
	ExpirationDateTime	string
}

// checkRequest Структура запроса
type qiwiCheckRequest struct {
	Bill	bill
}

// getApiSignatureFromCtx берет из контекста заголовок api_signature для авторизации
func getApiSignatureFromCtx(ctx *gin.Context) string {
	return ctx.Request.Header.Get("X-Api-Signature-SHA256")
}

// QiwiCheck Проверяет сообщение с сервера qiwi p2p об оплате счета
func (h *PayHandler) QiwiNotify(ctx *gin.Context) {

	var req qiwiCheckRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Парсим параметры из тела и заголовка уведомления
	params := strings.Join([]string{req.Bill.Amount.Currency, fmt.Sprint(req.Bill.Amount.Value), req.Bill.BillID, req.Bill.SiteId, req.Bill.Status.Value}, "|")
	api_signature := getApiSignatureFromCtx(ctx)

	// Авторизируем уведомление через middleware
	ctx.Set("p2pParams", params)
	ctx.Set("api_signature", api_signature)
	ctx.Next()

	// Если счет уже не в статусе ожидания, обрабатываем результат счета
	if req.Bill.Status.Value != "WAITING" {
		h.qiwiService.NotifyFromQiwi(ctx, req.Bill.Status.Value, req.Bill.BillID)
	}

	// Отправляем серверу qiwi ответ об успешной обработке уведомления
	ctx.JSON(http.StatusOK, gin.H{
		"error": "0",
	})

	return
}

