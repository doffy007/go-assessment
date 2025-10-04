package auth

import (
	"crypto/rsa"
	"os"
	"rest-api/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	jwtCfg := config.AppConfig.JWT

	privData, err := os.ReadFile(jwtCfg.JWTSecretPrivateKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read private key")
	}

	pubData, err := os.ReadFile(jwtCfg.JWTSecretPublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read public key")
	}

	PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privData)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse private key")
	}

	PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubData)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse public key")
	}

	log.Info().
		Str("publicKeyPath", jwtCfg.JWTSecretPublicKeyPath).
		Str("privateKeyPath", jwtCfg.JWTSecretPrivateKeyPath).
		Msg("initialized auth")
}
