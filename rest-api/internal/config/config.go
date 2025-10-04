package config

import (
	"os"
	"rest-api/internal/constant"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var AppConfig *DefaultConfig

func ConfigApps() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to load .env file")
	}

	envType := viper.GetString("ENV_TYPE")
	if envType == "" || !isValidEnvironmentType(constant.EnvironmentType(envType)) {
		log.Fatal().Msg("Invalid or missing ENV_TYPE. Valid values are [dev | stag | prod | test]")
	}

	conf := DefaultConfig{
		Apps: Apps{
			Name:           viper.GetString("APP_NAME"),
			Version:        viper.GetString("APP_VERSION"),
			RequestTimeout: viper.GetDuration("REQUEST_TIMEOUT"),
		},
		Server: Server{
			Port:    viper.GetInt("SERVER_PORT"),
			Zerolog: viper.GetInt("ZERO_LOG"),
		},
		Database: Database{
			Postgres: Datasource{
				Url:               viper.GetString("DATABASE_URL"),
				Port:              viper.GetInt("DATABASE_PORT"),
				DatabaseName:      viper.GetString("DATABASE_NAME"),
				Username:          viper.GetString("DATABASE_USERNAME"),
				Password:          viper.GetString("DATABASE_PASSWORD"),
				Schema:            viper.GetString("DATABASE_SCHEMA"),
				ConnectionTimeout: viper.GetDuration("DATABASE_CONNECTION_TIMEOUT"),
				MaxIdleConnection: viper.GetInt("DATABASE_MAX_IDLE_CONNECTION"),
				MaxOpenConnection: viper.GetInt("DATABASE_MAX_OPEN_CONNECTION"),
				DebugMode:         viper.GetBool("DATABASE_DEBUG_MODE"),
				DSN:               viper.GetString("DATABASE_DSN"),
				PingInterval:      viper.GetInt64("DATABASE_PING_INTERVAL"),
			},
		},
		Sonyflake: Sonyflake{
			IP: viper.GetString("SONYFLAKE_IP"),
		},
		Email: Email{
			SuspiciousEmail: viper.GetBool("SUSPICIOUS_EMAIL_DETECTION_ENABLE"),
		},
		JWT: JWT{
			JWTSecretPublicKeyPath:  viper.GetString("JWT_PUBLIC_KEY"),
			JWTSecretPrivateKeyPath: viper.GetString("JWT_PRIVATE_KEY"),
			ProjectName:             viper.GetString("PROJECT_NAME"),
		},
	}

	// Set default values if not provided
	if conf.Server.Port == 0 {
		conf.Server.Port = 8080
	}
	if conf.Database.Postgres.ConnectionTimeout == 0 {
		conf.Database.Postgres.ConnectionTimeout = 30 * time.Second
	}
	if conf.Database.Postgres.MaxIdleConnection == 0 {
		conf.Database.Postgres.MaxIdleConnection = 10
	}
	if conf.Database.Postgres.MaxOpenConnection == 0 {
		conf.Database.Postgres.MaxOpenConnection = 100
	}
	if conf.Database.Postgres.PingInterval == 0 {
		conf.Database.Postgres.PingInterval = 15
	}

	AppConfig = &conf
}

func isValidEnvironmentType(envType constant.EnvironmentType) bool {
	switch envType {
	case constant.DEV, constant.STAG, constant.PROD, constant.TEST:
		return true
	}
	return false
}
