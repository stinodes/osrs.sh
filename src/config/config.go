package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	Port   string `default:"5250"`
	Host   string `default:"localhost"`
	IDFile string `default:".ssh/id_ed25519"`
}

func LoadAppConfig() (c AppConfig, err error) {
	if err = envconfig.Process("OSRS", &c); err != nil {
		log.Fatal().Msgf("Unable to decode env.")
		return
	}

	return
}
