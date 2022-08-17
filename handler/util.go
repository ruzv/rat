package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Wrap(handler func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler(c)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": err.Error(),
				},
			)
		}
	}
}
