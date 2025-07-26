package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GenRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := uuid.New()
		c.Set("reqId", reqId.String())
		c.Next()
	}
}
