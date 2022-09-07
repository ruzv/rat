package handler

import (
	"net/http"

	"private/rat/errors"
	"private/rat/logger"

	"github.com/gin-gonic/gin"
)

func Wrap(handler func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler(c)
		if err != nil {
			logger.Errorf("handler error: %v", err)
		}
	}
}

// WriteErrorJSON writes code and error message in JSON format as http response.
func WriteErrorJSON(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"code": code, "message": message})
}

func Body[T any](c *gin.Context) (T, error) { //nolint:ireturn
	var body, empty T

	err := c.BindJSON(&body)
	if err != nil {
		WriteErrorJSON(c, http.StatusBadRequest, "failed to read body")

		return empty, errors.Wrap(err, "failed to bind body struct")
	}

	return body, nil
}
