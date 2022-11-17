package handlers

import (
	"fmt"
	"strings"
	"net/http"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// customFields Дополнительные данные счета (нет)
type customFields struct {}

// customer Данные о пользователе, на которого был выставлен счет
type customer struct {
	Phone	string	`json:"phone"`
	Email	string	`json:"email"`
	Account	string	`json:"account"`
}

// status Данные о статусе счета
type status struct {
	Value			string	`json:"value"`
	ChangedDatetime	string	`json:"datetime"`
}

// amount Данные о сумме счета
type amount struct {
	Value		float64	`json:"value"`
	Currency	string	`json:"currency"`
}

// bill Данные о счете
type bill struct {
	SiteId				string			`json:"siteId"`
	BillID				string			`json:"billId"`
	Amount				amount			`json:"amount"`
	Status				status			`json:"status"`
	Customer			customer		`json:"customer"`
	CustomFields		customFields	`json:"customFields"`
	CreationDateTime	string			`json:"creationDateTime"`
	ExpirationDateTime	string			`json:"expirationDateTime"`
}

// checkRequest Структура запроса
type qiwiCheckRequest struct {
	Bill	bill	`json:"bill"`
	Version	string	`json:"version"`
}

// getApiSignatureFromCtx берет из контекста заголовок api_signature для авторизации
func getApiSignatureFromCtx(ctx *gin.Context) string {
	return ctx.Request.Header.Get("X-Api-Signature-SHA256")
}

// QiwiCheck Проверяет сообщение с сервера qiwi p2p об оплате счета
func (h *PayHandler) QiwiNotify(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "QiwiNotifyHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("QiwiNotifyHandler() started")
	defer l.Debug("QiwiNotifyHandler() done")

	var req qiwiCheckRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, l, &req, "POST", MetricQiwiNotify); !ok {
		return
	}

	// Парсим параметры из тела и заголовка уведомления
	params := strings.Join([]string{req.Bill.Amount.Currency, fmt.Sprint(req.Bill.Amount.Value), req.Bill.BillID, req.Bill.SiteId, req.Bill.Status.Value}, "|")
	api_signature := getApiSignatureFromCtx(ctx)

	// Авторизируем уведомление
	if QiwiAuthorization(ctx, l, h.key, params, api_signature) {
		// Если счет уже не в статусе ожидания, обрабатываем результат счета
		if req.Bill.Status.Value != "WAITING" {
			if err := h.qiwiService.NotifyFromQiwi(ctx, req.Bill.Status.Value, req.Bill.BillID); err != nil {
				InternalErrResp(ctx, l, err)

				Bridge(ctx, http.StatusInternalServerError, "POST", MetricQiwiNotify)
				
				return
			}
		}

		// Отправляем серверу qiwi ответ об успешной обработке уведомления
		ctx.JSON(http.StatusOK, gin.H{
			"error": "0",
		})

		Bridge(ctx, http.StatusOK, "POST", MetricQiwiNotify)
	}

	return
}
