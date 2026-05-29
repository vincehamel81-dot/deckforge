package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CorrelationIDHeader = "X-Correlation-ID"
const CorrelationIDKey = "correlationID"

func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cid := c.GetHeader(CorrelationIDHeader)
		if cid == "" {
			cid = uuid.New().String()
		}
		c.Set(CorrelationIDKey, cid)
		c.Header(CorrelationIDHeader, cid)
		c.Next()
	}
}

func CorrelationID(c *gin.Context) string {
	val, _ := c.Get(CorrelationIDKey)
	s, _ := val.(string)
	return s
}
