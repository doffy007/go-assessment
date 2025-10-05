package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CSVPath string `envconfig:"CSV_PATH" default:"sample"`
	Workers int    `envconfig:"WORKERS" default:"4"`
	Port    int    `envconfig:"PORT" default:"8080"`
}

var AppConfig Config

func init() {
	if err := envconfig.Process("", &AppConfig); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
}
