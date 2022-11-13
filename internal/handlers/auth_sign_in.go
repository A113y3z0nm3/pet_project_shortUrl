package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"short_url/internal/models"
	myLog "short_url/pkg/logger"
)

// Структура запроса
type signInRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,gte=6,lte=30"`
}

// Структура ответа
type signInResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

// SignIn метод AuthService для выполнения входа
func (h *AuthHandler) SignIn(ctx *gin.Context) {
	ctxLog := myLog.ContextWithSpan(ctx, "SignIn")
	l := h.logger.WithContext(ctxLog)

	l.Debug("SignIn() started")
	defer l.Debug("SignIn() done")

	var req signInRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Ищем пользователя и пароль в базе
	u, err := h.authService.SignInUserByName(ctxLog, models.SignInUserDTO{
		Username: req.Username,
		Password: req.Password,
	})

	// Обрабатываем ошибки
	if err != nil {
		log.Println(err)

		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user with username=%s not found", req.Username),
			})

			return
		}

		if err.Error() == "invalid password" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid password",
			})

			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})

		return
	}

	// Создаем токен доступа
	token, err := h.tokenService.CreateToken(ctxLog, models.CreateTokenDTO{
		Username:	u.Username,
		Subscribe:	u.Subscribe,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})

		return
	}

	// Маппим данные в ответ
	ctx.JSON(http.StatusOK, signInResponse{
		AccessToken: token,
		Username:    u.Username,
	})

	return
}