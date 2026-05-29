package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
)

// DealerMiddleware verifies the authenticated user is the dealer of the requested game.
// Must run after AuthMiddleware. Injects the loaded game into context under "game".
func DealerMiddleware(games game.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFromContext(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		gameID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
			return
		}
		g, err := games.FindByID(gameID)
		if err != nil || g == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
		if g.DealerUserID.String() != claims.UserID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only the dealer can perform this action"})
			return
		}
		c.Set("game", g)
		c.Next()
	}
}
