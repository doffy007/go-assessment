package uid

import (
	"os"
	"rest-api/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sony/sonyflake"
)

type uidConfig struct {
	IP string
}

var (
	cfg uidConfig
	sf  *sonyflake.Sonyflake
)

func init() {
	config.ConfigApps()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = uidConfig{
		IP: config.AppConfig.Sonyflake.IP,
	}

	sf = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: machineID,
	})

	if sf == nil {
		log.Fatal().Msg("failed to initialize sonyflake")
	}

	log.Info().
		Interface("config", cfg).
		Msg("initialize uid")
}
