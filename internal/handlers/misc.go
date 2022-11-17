package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"short_url/internal/handlers/middlewares"
	"short_url/internal/models"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Имена обработчика для контекста middleware
const (
	MetricSignIn		= "signIn"
	MetricSignUp		= "signUp"

	MetricQiwiNotify	= "qiwiNotify"
	MetricQiwiSubExt	= "qiwiSubExt"
	MetricQiwiSub		= "qiwiSub"

	MetricCreateLink	= "createLink"
	MetricCreateQR		= "createQR"
	MetricDeleteLink	= "deleteLink"
	MetricGetAllLinks	= "getAllLinks"
	MetricGetLink		= "getLink"
	MetricRedirectLink	= "redirectLink"
)

// hash256 Создает новый хеш
func hash256() hash.Hash {
	return sha256.New()
}

// Bridge Мост к middleware, передающие значения по лейблам метрики в контекст
func Bridge(ctx *gin.Context, code int, method string, handler string) {
	ctx.Set(middlewares.Code, fmt.Sprint(code))
	ctx.Set(middlewares.Method, method)
	ctx.Set(middlewares.Handler, handler)

	return
}

// InternalErrResp Вспомогательная функция (записывает внутреннюю ошибку в лог, возвращает ответ "ошибка сервера")
func InternalErrResp(ctx *gin.Context, l *log.Log, err error) {
	l.Errorf("Internal server error: %s", err)

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})

	return
}

// GetUserInfo Возвращает информацию о пользователе и его подписке из контекста
func GetUserInfo(ctx *gin.Context) (models.JWTUserInfo, error) {
	// Получаем данные из контекста
	info, ok := ctx.Get(middlewares.UserInfo)
	if !ok {

		return models.JWTUserInfo{}, errors.New("failed to get user info")
	}

	// Преобразуем их в структуру для ответа
	user, ok := info.(models.JWTUserInfo)
	if !ok {

		return models.JWTUserInfo{}, errors.New("failed to JWT user info")
	}

	return user, nil
}

// QiwiAuthorizeHash Авторизирует входящее уведомление о счете
func QiwiAuthorization(ctx *gin.Context, l *log.Log, key string, params, sign string) bool {
	// Создаем хеш с применением ключа
	hash := hmac.New(hash256, []byte(key))

	// Хешируем параметры
	_, err := hash.Write([]byte(params))
	if err != nil {
		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "POST", MetricQiwiNotify)

		return false
	}

	// Сравниваем полученную строку с заголовком
	str := string(hash.Sum(nil))
	if sign != str {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authorization failed",
		})

		Bridge(ctx, http.StatusUnauthorized, "POST", MetricQiwiNotify)

		return false
	}

	return true
}
