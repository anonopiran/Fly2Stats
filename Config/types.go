package config

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"
)

type PathType string
type RedisUrlType string
type InfluxUrlType string
type V2rayUrlType string
type LogLevelType string

type SettingsType struct {
	InfluxdbUrl     InfluxUrlType     `env:"INFLUXDB_URL" env-required:"true"`
	V2flyApiAddress []V2rayUrlType    `env:"V2FLY_API_ADDRESS" env-required:"true"`
	InfluxdbTags    map[string]string `env:"INFLUXDB_TAGS" env-default:""`
	RedisUrl        RedisUrlType      `env:"REDIS_URL" env-default:""`
	CheckpointPath  PathType          `env:"CHECKPOINT_PATH" env-default:"./storage/checkpoints"`
	UpdateInterval  int               `env:"UPDATE_INTERVAL" env-default:"5"`
	LogLevel        LogLevelType      `env:"LOG_LEVEL" env-default:"warning"`
}

func (f *PathType) AsString() string {
	return string(*f)
}
func (f *RedisUrlType) AsOpts() redis.Options {
	opt, _ := redis.ParseURL(string(*f))
	return *opt
}
func (f *InfluxUrlType) AsUrl() url.URL {
	u, _ := url.Parse(string(*f))
	return *u
}
func (f *V2rayUrlType) AsUrl() url.URL {
	u, _ := url.Parse(string(*f))
	return *u
}

// .............
func (f *PathType) SetValue(s string) error {
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		log.WithField("path", f).WithError(err).Error("can not create dir")
		return err
	}
	*f = PathType(s)
	return nil
}
func (f *LogLevelType) SetValue(s string) error {
	ll, err := log.ParseLevel(s)
	if err != nil {
		log.WithField("value", s).WithError(err).Error("error while parsing config")
		return err
	}
	log.SetLevel(ll)
	// log.SetReportCaller(true)
	*f = LogLevelType(s)
	return nil
}
func (f *RedisUrlType) SetValue(s string) error {
	if s == "" {
		log.Warning("REDIS_URL not defined")

	} else {
		_, err := redis.ParseURL(s)
		if err != nil {
			log.WithField("value", s).WithError(err).Error("error while parsing config")
			return err
		}
		*f = RedisUrlType(s)
	}
	return nil
}
func (f *InfluxUrlType) SetValue(s string) error {
	LogWithRaw := log.WithField("value", s)
	u, err := url.Parse(s)
	if err != nil {
		LogWithRaw.WithError(err).Error("error while parsing config")
		return err
	}
	if u.Host == "" || u.Scheme == "" {
		e := errors.New("scheme of host not provided")
		LogWithRaw.Error(e)
		return e
	}
	if u.User.Username() == "" {
		e := errors.New("org not provided")
		LogWithRaw.Error(e)
		return e
	}
	token, _ := u.User.Password()
	if token == "" {
		e := errors.New("token not provided")
		LogWithRaw.Error(e)
		return e
	}
	if strings.Trim(u.Path, "/") == "" {
		e := errors.New("bucket not provided")
		LogWithRaw.Error(e)
		return e
	}
	*f = InfluxUrlType(s)
	return nil
}
func (f *V2rayUrlType) SetValue(s string) error {
	LogWithRaw := log.WithField("value", s)
	u, err := url.Parse(s)
	if err != nil {
		LogWithRaw.WithError(err).Error("error while parsing config")
		return err
	}
	if u.Hostname() == "" {
		e := errors.New("hostname not provided")
		LogWithRaw.Error(e)
		return e
	}
	if u.Port() == "" {
		e := errors.New("port not provided")
		LogWithRaw.Error(e)
		return e
	}
	*f = V2rayUrlType(s)
	return nil
}
