package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"
)

// Структура запроса
type signUpRequest struct {
	Username	string				`json:"username" binding:"required"`
	FirstName	string				`json:"first_name" binding:"required"`
	LastName	string				`json:"last_name" binding:"required"`
	Password	string				`json:"password" binding:"required,gte=6,lte=30"`
}

// SignUp метод AuthService для регистрации пользователя
func (h *AuthHandler) SignUp(ctx *gin.Context) {
	ctxLog := myLog.ContextWithSpan(ctx, "SignUp")
	l := h.logger.WithContext(ctxLog)

	l.Debug("SignUp() started")
	defer l.Debug("SignUp() done")

	var req signUpRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Регистрируем пользователя
	err := h.authService.SignUpUser(ctxLog, models.SignUpUserDTO{
		Username:	req.Username,
		Password:	req.Password,
		LastName:	req.LastName,
		FirstName:	req.FirstName,
	})

	// Обрабатываем ошибки
	if err != nil {
		if err.Error() == "user exist" {
			ctx.JSON(http.StatusConflict, gin.H{
				"Error": fmt.Sprintf("user with username=%s exist", req.Username),
			})

			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Internal server error",
		})

		return
	}

	ctx.JSON(http.StatusOK, "OK")

	return
}
