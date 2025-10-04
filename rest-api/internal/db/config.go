package db

import (
	"os"
	"rest-api/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type dbConfig struct {
	DSN          string
	PingInterval int64
}

var cfg dbConfig

func init() {
	config.ConfigApps()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = dbConfig{
		DSN:          config.AppConfig.Database.Postgres.DSN,
		PingInterval: config.AppConfig.Database.Postgres.PingInterval,
	}

	log.Info().
		Interface("config", cfg).
		Msg("initialize postgres")
}
