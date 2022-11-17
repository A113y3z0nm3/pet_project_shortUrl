package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// authHeader Структура для заголовка авторизации
type authHeader struct {
	Token string `header:"Authorization"`
}

// metricSet Отправляет метрики из middleware, если по каким-то ошибкам был пропущен основной обработчик
func metricSet(ctx *gin.Context, code int) {
	ctx.Set(Handler, MidAuth)
	ctx.Set(Method, MidAuth)
	ctx.Set(Code, fmt.Sprint(code))
	ctx.Set(Skip, "true")
}

// AuthUser извлекает пользователя из заголовка Authorization.
// Устанавливает пользователя в контекст, если пользователь существует
func (m *Middlewares) AuthUser(ctx *gin.Context) {

	h := authHeader{}

	// bind Authorization Header to h and check for validation errors
	if err := ctx.ShouldBindHeader(&h); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "bad header values",
		})

		metricSet(ctx, http.StatusBadRequest)

		return
	}

	tokenHeader := strings.Split(h.Token, "Bearer ")
	if len(tokenHeader) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "must provide Authorization header with format `Bearer {token}`",
		})

		metricSet(ctx, http.StatusBadRequest)

		return
	}

	// validate ID token here
	info, err := m.tokenService.ValidateToken(ctx, tokenHeader[1])
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "provided token is invalid",
		})

		metricSet(ctx, http.StatusBadRequest)

		return
	}

	ctx.Set(UserInfo, info)

	ctx.Next()
}
