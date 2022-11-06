package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// authHeader Структура для заголовка авторизации
type authHeader struct {
	Token string `header:"Authorization"`
}

// AuthUser извлекает пользователя из заголовка Authorization.
// Устанавливает пользователя в контекст, если пользователь существует
func (m *Middlewares) AuthUser(ctx *gin.Context) {

	h := authHeader{}

	// bind Authorization Header to h and check for validation errors
	if err := ctx.ShouldBindHeader(&h); err != nil {
		m.log.Error("middleware: bad auth-header values")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "bad header values",
		})

		ctx.Abort()
		return
	}

	tokenHeader := strings.Split(h.Token, "Bearer ")
	if len(tokenHeader) < 2 {
		m.log.Error("middleware: invalid token header format")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "must provide Authorization header with format `Bearer {token}`",
		})

		ctx.Abort()
		return
	}

	// validate ID token here
	info, err := m.tokenService.ValidateToken(ctx, tokenHeader[1])
	if err != nil {
		m.log.Error("middleware: invalid auth token")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "provided token is invalid",
		})

		ctx.Abort()
		return
	}

	ctx.Set(UserInfo, info)

	ctx.Next()
}
