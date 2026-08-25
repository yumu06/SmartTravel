package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 200, "data": fmt.Sprint(err), "msg": nil})
			}
		}()

		c.Next()
	}
}
