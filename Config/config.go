package config

import (
	"fmt"
	"os"

	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}
var settingsInstance *SettingsType

func Describe() {
	var cfg SettingsType
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}

func Config() *SettingsType {
	if settingsInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		settingsInstance = &SettingsType{}
		var err error = nil
		if _, err_file := os.Stat(".env"); err_file == nil {
			err = cleanenv.ReadConfig(".env", settingsInstance)
			log.Info("found .env file")
		} else {
			err = cleanenv.ReadEnv(settingsInstance)
			log.Info("no .env file found")
		}
		if err != nil {
			log.WithError(err).Panic("can not initiate configuration")
		}
		log.WithField("data", fmt.Sprintf("%+v", settingsInstance)).Debug("Parsed Configuration")
	}
	return settingsInstance
}
