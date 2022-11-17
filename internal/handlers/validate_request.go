package handlers

import (
	"net/http"
	log "short_url/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// used to help extract validation errors
type invalidArgument struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Tag   string      `json:"tag"`
	Param string      `json:"param"`
}

// bindData is helper function, returns false if data is not bound
func bindData(c *gin.Context, l *log.Log, req interface{}, method, handler string) bool {
	if c.ContentType() != "application/json" {

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Endpoint only accepts Content-Type application/json",
		})

		Bridge(c, http.StatusUnsupportedMediaType, method, handler)

		return false
	}
	// Bind incoming json to struct and check for validation errors
	if err := c.ShouldBind(req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			// could probably extract this, it is also in middleware_auth_user
			var invalidArgs []invalidArgument

			for _, err := range errs {
				invalidArgs = append(invalidArgs, invalidArgument{
					err.Field(),
					err.Value(),
					err.Tag(),
					err.Param(),
				})
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "Invalid request parameters. See invalidArgs",
				"invalidArgs": invalidArgs,
			})

			Bridge(c, http.StatusBadRequest, method, handler)

			return false
		}
		
		l.Errorf("Error binding data: %+v\n", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})

		Bridge(c, http.StatusInternalServerError, method, handler)

		return false
	}

	return true
}
