// @title DeckForge API
// @version 1.0
// @description Card-game engine: shoe management, dealing, shuffling, scoring.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT obtained from POST /auth/register or POST /auth/login. Enter the full value including the prefix: Bearer &lt;your-token&gt;

package main

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vincehamel81-dot/deckforge/config"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
	"github.com/vincehamel81-dot/deckforge/internal/infrastructure/persistence"
	ws "github.com/vincehamel81-dot/deckforge/internal/infrastructure/ws"
	httpserver "github.com/vincehamel81-dot/deckforge/internal/presentation/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	_ = godotenv.Load() // load .env if present; ignore error if not found

	cfg := config.Load()

	// Log level — change LOG_LEVEL env var and restart; no deploy needed.
	if level, err := zerolog.ParseLevel(cfg.LogLevel); err == nil {
		zerolog.SetGlobalLevel(level)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Log format — "text" gives human-readable output for local dev;
	// "json" is the default for production log aggregators (Railway, Datadog…).
	if cfg.LogFormat == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if cfg.JWTSecret == "" {
		log.Fatal().Msg("JWT_SECRET must be set")
	}

	db, err := persistence.NewDB(cfg.DBDriver, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Str("db", cfg.DatabaseURL).Msg("database connected")

	games := persistence.NewGameRepo(db)
	players := persistence.NewPlayerRepo(db)
	shoes := persistence.NewShoeRepo(db)
	users := persistence.NewUserRepo(db)

	for _, username := range cfg.AdminSeedUsernames {
		exists, err := users.ExistsByUsername(username)
		if err != nil {
			log.Fatal().Err(err).Str("username", username).Msg("failed to check admin seed user")
		}
		if !exists {
			admin := user.New(username)
			admin.Role = user.RoleAdmin
			if err := users.Create(admin); err != nil {
				log.Fatal().Err(err).Str("username", username).Msg("failed to create admin seed user")
			}
			log.Info().Str("username", username).Msg("admin user seeded")
		}
	}

	// Hub is declared first so the disconnect callback can close over it.
	// The callback fires only after DISCONNECT_TIMEOUT_SECONDS (≥30s), well
	// after hub is assigned below.
	var hub *ws.Hub

	disconnectTimeout := time.Duration(cfg.DisconnectTimeoutSeconds) * time.Second
	onDisconnect := func(gameID, userID string) {
		gid, err := uuid.Parse(gameID)
		if err != nil {
			return
		}
		uid, err := uuid.Parse(userID)
		if err != nil {
			return
		}

		p, err := players.FindByUserAndGame(uid, gid)
		if err != nil || p == nil || !p.IsActive() {
			return // player already left or game gone
		}

		gameEnded, err := commands.RemovePlayer(commands.RemovePlayerCommand{
			GameID:          gid,
			PlayerID:        p.ID,
			RequesterUserID: uid,
			IsAdmin:         false,
		}, games, players, shoes)
		if err != nil {
			log.Warn().Str("gameId", gameID).Str("userId", userID).Err(err).
				Msg("disconnect auto-remove failed")
			return
		}

		log.Info().
			Str("gameId", gameID).
			Str("userId", userID).
			Int("timeoutSeconds", cfg.DisconnectTimeoutSeconds).
			Msg("player auto-removed after WS disconnect timeout")

		hub.Broadcast(gameID, ws.Message{
			Event:   ws.EventPlayerLeft,
			Payload: map[string]interface{}{"userId": userID},
		})
		if gameEnded {
			hub.Broadcast(gameID, ws.Message{
				Event:   ws.EventGameEnded,
				Payload: map[string]interface{}{"reason": "not_enough_players"},
			})
		}
	}

	hub = ws.NewHub(disconnectTimeout, onDisconnect)
	router := httpserver.NewRouter(cfg, games, players, shoes, users, hub)

	log.Info().Str("port", cfg.Port).Msg("DeckForge listening")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
