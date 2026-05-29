package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vincehamel81-dot/deckforge/config"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg := config.Load()

	log.Info().
		Str("port", cfg.Port).
		Str("db", cfg.DatabaseURL).
		Msg("DeckForge starting")

	// router wired in commit 9
	// db wired in commit 3
	select {}
}
