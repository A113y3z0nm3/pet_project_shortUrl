package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"net/http"

	"github.com/gin-gonic/gin"
)

// hash256 Создает новый хеш
func hash256() hash.Hash {
	return sha256.New()
}

// QiwiAuthorizeHash Авторизирует входящее уведомление о счете
func (m *Middlewares) QiwiAuthorization(ctx *gin.Context) {
	// Извлекаем параметры запроса из контекста
	ctxparams, ok := ctx.Get("p2pParams")
	if !ok {
		// log

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse req params",
		})

		ctx.Abort()

		return
	}
	params := ctxparams.(string)

	// Извлекаем заголовок из контекста
	ctxheader, ok := ctx.Get("api_signature")
	if !ok {
		// log

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse api_signature",
		})

		ctx.Abort()

		return
	}
	header := ctxheader.(string)

	// Создаем хеш с применением ключа
	hash := hmac.New(hash256, []byte(m.key))

	// Хешируем параметры
	_, err := hash.Write([]byte(params))
	if err != nil {
		// log

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		ctx.Abort()

		return
	}

	// Сравниваем полученную строку с заголовком
	str := string(hash.Sum(nil))
	if header != str {
		// log

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "authorization failed",
		})

		ctx.Abort()

		return
	}

	return
}