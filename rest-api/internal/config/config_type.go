package config

import (
	"time"
)

type DefaultConfig struct {
	Apps      Apps      `mapstructure:"apps"`
	Server    Server    `mapstructure:"server"`
	Database  Database  `mapstructure:"database"`
	Sonyflake Sonyflake `mapstructure:"sonyflake"`
	Email     Email     `mapstructure:"email"`
	JWT       JWT       `mapstructure:"jwt"`
}

type Apps struct {
	Name           string        `mapstructure:"name"`
	Version        string        `mapstructure:"version"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout"`
}

type Server struct {
	Port    int `mapstructure:"port"`
	Zerolog int `mapstructure:"zerolog"`
}

type Datasource struct {
	Url               string        `mapstructure:"url"`
	Port              int           `mapstructure:"port"`
	DatabaseName      string        `mapstructure:"databaseName"`
	Username          string        `mapstructure:"username"`
	Password          string        `mapstructure:"password"`
	Schema            string        `mapstructure:"schema"`
	ConnectionTimeout time.Duration `mapstructure:"connectionTimeout"`
	MaxIdleConnection int           `mapstructure:"maxIdleConnection"`
	MaxOpenConnection int           `mapstructure:"maxOpenConnection"`
	DebugMode         bool          `mapstructure:"debugMode"`
	DSN               string        `mapstructure:"databaseDSN"`
	PingInterval      int64         `mapstructure:"pingInterval"`
}

type Database struct {
	Postgres Datasource `mapstructure:"postgres"`
}

type JWT struct {
	JWTSecretPublicKeyPath  string `mapstructure:"jwtSecretPublicKeyPath"`
	JWTSecretPrivateKeyPath string `mapstructure:"jwtSecretPrivateKeyPath"`
	ProjectName             string `mapstructure:"projectName"`
}

type Sonyflake struct {
	IP string `mapstructure:"sonyflake_ip"`
}

type Email struct {
	SuspiciousEmail bool `mapstructure:"suspicious_email_detection_enable"`
}
