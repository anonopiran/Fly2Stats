package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

type SettingsType struct {
	V2fly_Api_Address string `env:"V2FLY_API_ADDRESS" env-required:"true"`
	Influxdb_Url      string `env:"INFLUXDB_URL" env-required:"true"`
	Influxdb_Org      string `env:"INFLUXDB_ORG" env-required:"true"`
	Influxdb_Token    string `env:"INFLUXDB_TOKEN" env-required:"true"`
	Influxdb_Bucket   string `env:"INFLUXDB_BUCKET" env-required:"true"`
	Checkpoint_Path   string `env:"CHECKPOINT_PATH" env-default:"./storage/checkpoints"`
	Update_Interval   int    `env:"UPDATE_INTERVAL" env-default:"5"`
	Log_Level         string `env:"LOG_LEVEL" env-default:"warning"`
}

func Config() SettingsType {
	var cfg SettingsType
	// load_dot_env()
	var err error = nil
	if _, err = os.Stat(".env"); err == nil {
		err = cleanenv.ReadConfig(".env", &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}
	if err != nil {
		log.WithError(err).Fatalln("can not initiate configuration")
	}
	ll, err := log.ParseLevel(cfg.Log_Level)
	if err != nil {
		log.WithError(err).Error("can not set log level")
		ll = log.GetLevel()
	}
	log.SetLevel(ll)
	log.WithField("data", fmt.Sprintf("%+v", cfg)).Debug("Parsed Configuration")
	err = os.MkdirAll(cfg.Checkpoint_Path, os.ModePerm)
	if err != nil {
		log.WithError(err).Panic("can not create checkpoint dir")
	}
	return cfg
}
func Describe() {
	var cfg SettingsType
	help, err := cleanenv.GetDescription(&cfg, nil)
	if err != nil {
		log.WithError(err).Panic("can not generate description")
	}
	log.Println(help)
}
