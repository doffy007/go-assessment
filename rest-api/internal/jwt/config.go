package jwt

import (
	"io/ioutil"
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

var (
	cfg              jwtConfig
	jwtSecretPublic  []byte
	jwtSecretPrivate []byte
)

func init() {
	config.ConfigApps()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = jwtConfig{
		JWTSecretPublicKeyPath:  config.AppConfig.JWT.JWTSecretPublicKeyPath,
		JWTSecretPrivateKeyPath: config.AppConfig.JWT.JWTSecretPrivateKeyPath,
		ProjectName:             config.AppConfig.JWT.ProjectName,
	}

	var err error
	jwtSecretPublic, err = ioutil.ReadFile(cfg.JWTSecretPublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading public key file")
	}

	jwtSecretPrivate, err = ioutil.ReadFile(cfg.JWTSecretPrivateKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading private key file")
	}

	if len(jwtSecretPublic) == 0 || len(jwtSecretPrivate) == 0 {
		log.Fatal().Msg("JWT secret keys are not properly configured")
	}
}
