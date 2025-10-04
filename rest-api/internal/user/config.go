package user

import (
	"os"
	"regexp"
	"rest-api/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type userConfig struct {
	SuspiciousEmailDetectionEnable bool
	SuspiciousEmailRegex           *regexp.Regexp
}

var cfg *userConfig

func init() {
	config.ConfigApps()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = &userConfig{
		SuspiciousEmailDetectionEnable: config.AppConfig.Email.SuspiciousEmail,
	}

	if cfg.SuspiciousEmailDetectionEnable {
		cfg.SuspiciousEmailRegex = regexp.MustCompile(`(?i)spam|fake|test|temp|mailinator`)
	}

	log.Info().
		Bool("SuspiciousEmailDetectionEnable", cfg.SuspiciousEmailDetectionEnable).
		Msg("initialize suspicious email detection")
}
