package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"short_url/internal/models"
	log "short_url/pkg/logger"
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
	ctxLog := log.ContextWithSpan(ctx, "SignUpHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("SignUpHandler() started")
	defer l.Debug("SignUpHandler() done")

	var req signUpRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, l, &req, "POST", MetricSignUp); !ok {
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

			Bridge(ctx, http.StatusConflict, "POST", MetricSignUp)

			return
		}

		InternalErrResp(ctx, l, err)

		Bridge(ctx, http.StatusInternalServerError, "POST", MetricSignUp)

		return
	}

	ctx.JSON(http.StatusOK, "OK")

	Bridge(ctx, http.StatusOK, "POST", MetricSignUp)

	return
}
