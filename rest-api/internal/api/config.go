package api

import (
	"os"
	"rest-api/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type jwtConfig struct {
	JWTSecretPublicKeyPath  string
	JWTSecretPrivateKeyPath string
	ProjectName             string
}

var cfg jwtConfig

func init() {
	config.ConfigApps()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = jwtConfig{
		JWTSecretPublicKeyPath:  config.AppConfig.JWT.JWTSecretPublicKeyPath,
		JWTSecretPrivateKeyPath: config.AppConfig.JWT.JWTSecretPrivateKeyPath,
		ProjectName:             config.AppConfig.JWT.ProjectName,
	}

	log.Info().
		Interface("config", cfg).
		Msg("initialize api middleware")
}
