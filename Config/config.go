package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

func Describe() {
	var cfg SettingsType
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}

func Config() SettingsType {
	var cfg SettingsType
	var err error = nil
	if _, err_file := os.Stat(".env"); err_file == nil {
		err = cleanenv.ReadConfig(".env", &cfg)
		log.Info("found .env file")
	} else {
		err = cleanenv.ReadEnv(&cfg)
		log.Info("no .env file found")
	}
	if err != nil {
		log.WithError(err).Panic("can not initiate configuration")
	}
	log.WithField("data", fmt.Sprintf("%+v", cfg)).Debug("Parsed Configuration")
	return cfg
}
