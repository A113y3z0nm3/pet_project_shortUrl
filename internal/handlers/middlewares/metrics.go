package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
)

// Recorder получает из контекста значения для лейблов кастомной метрики запросов к эндпоинтам
func (m *Middlewares) Recorder(ctx *gin.Context) {

	ctx.Next()

	// Получаем значения метода запроса и кода ответа
	handler, ok := ctx.Get(Handler)

	if !ok {
		m.logger.Error("metrics: failed to get handler name from ctx")

		return
	}

	method, ok := ctx.Get(Method)

	if !ok {
		m.logger.Error("metrics: failed to get request method from ctx")

		return
	}

	code, ok := ctx.Get(Code)

	if !ok {
		m.logger.Error("metrics: failed to get response code from ctx")

		return
	}

	// Инкрементируем счетчик метрики
	m.Counter.With(prometheus.Labels{
		"method":  fmt.Sprint(method),
		"code":    fmt.Sprint(code),
		"handler": fmt.Sprint(handler),
	}).Inc()

	ctx.Abort()

	return
}
