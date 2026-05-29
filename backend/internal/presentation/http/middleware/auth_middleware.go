package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vincehamel81-dot/deckforge/internal/infrastructure/auth"
)

const UserClaimsKey = "userClaims"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.Validate(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(UserClaimsKey, claims)
		c.Next()
	}
}

// ClaimsFromContext extracts the authenticated user's claims from a Gin context.
func ClaimsFromContext(c *gin.Context) *auth.Claims {
	val, _ := c.Get(UserClaimsKey)
	claims, _ := val.(*auth.Claims)
	return claims
}
