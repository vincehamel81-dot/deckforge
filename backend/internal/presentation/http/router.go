package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincehamel81-dot/deckforge/config"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/handlers"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

func NewRouter(
	cfg *config.Config,
	games game.Repository,
	players player.Repository,
	shoes shoe.Repository,
	users user.Repository,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(corsMiddleware(cfg.CORSOrigin))
	r.Use(middleware.CorrelationMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())

	jwtExpiry, _ := time.ParseDuration(cfg.JWTExpiry)
	if jwtExpiry == 0 {
		jwtExpiry = 24 * time.Hour
	}

	authH := handlers.NewAuthHandler(users, cfg.JWTSecret, jwtExpiry)
	gameH := handlers.NewGameHandler(games, players, shoes)
	shoeH := handlers.NewShoeHandler(games, shoes)
	playerH := handlers.NewPlayerHandler(games, players, shoes)
	dealH := handlers.NewDealHandler(games, players, shoes, users)

	auth := middleware.AuthMiddleware(cfg.JWTSecret)
	dealer := middleware.DealerMiddleware(games)

	// Public routes
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/login", authH.Login)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/decks", shoeH.CreateDeck)

	// Authenticated game routes
	g := r.Group("/games", auth)
	{
		g.POST("", gameH.CreateGame)
		g.GET("", gameH.ListGames)
		g.GET("/:id", gameH.GetGame)
		g.DELETE("/:id", dealer, gameH.DeleteGame)
		g.POST("/:id/start", dealer, gameH.StartGame)
		g.POST("/:id/end", dealer, gameH.EndGame)

		g.POST("/:id/shoe/decks", dealer, shoeH.AddDeckToShoe)
		g.POST("/:id/shoe/shuffle", dealer, shoeH.ShuffleShoe)
		g.GET("/:id/shoe/suits", shoeH.GetSuitCounts)
		g.GET("/:id/shoe/cards", shoeH.GetCardCounts)

		g.POST("/:id/players", playerH.JoinGame)
		g.DELETE("/:id/players/:pid", playerH.LeaveGame)
		g.GET("/:id/players", dealH.GetLeaderboard)

		g.POST("/:id/players/:pid/deal", dealer, dealH.DealCards)
		g.GET("/:id/players/:pid/hand", dealH.GetPlayerHand)
	}

	return r
}

func corsMiddleware(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Correlation-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
