package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vincehamel81-dot/deckforge/config"
	"github.com/vincehamel81-dot/deckforge/internal/infrastructure/persistence"
	httpserver "github.com/vincehamel81-dot/deckforge/internal/presentation/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	_ = godotenv.Load() // load .env if present; ignore error if not found

	cfg := config.Load()

	if cfg.JWTSecret == "" {
		log.Fatal().Msg("JWT_SECRET must be set")
	}

	db, err := persistence.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Str("db", cfg.DatabaseURL).Msg("database connected")

	games := persistence.NewGameRepo(db)
	players := persistence.NewPlayerRepo(db)
	shoes := persistence.NewShoeRepo(db)
	users := persistence.NewUserRepo(db)

	router := httpserver.NewRouter(cfg, games, players, shoes, users)

	log.Info().Str("port", cfg.Port).Msg("DeckForge listening")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
