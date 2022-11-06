package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// BasicAuth Базовая авторизация
func (m *Middlewares) BasicAuth(ctx *gin.Context) {

	user, password, hasAuth := ctx.Request.BasicAuth()

	if !(hasAuth && len(user) != 0 && password == m.basicPassword) {
		m.log.Errorf("middleware: user with login=%s and password=%s unable auth", user, password)

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "user unauthorized",
		})

		ctx.Abort()
		return
	}

	ctx.Next()
}