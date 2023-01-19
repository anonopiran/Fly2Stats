package config

import (
	"errors"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type PathType string
type RabbitUrlType string
type InfluxUrlType string
type V2rayUrlType string
type LogLevelType string

type SettingsType struct {
	InfluxdbUrl     InfluxUrlType     `env:"INFLUXDB_URL" env-required:"true"`
	V2flyApiAddress []V2rayUrlType    `env:"V2FLY_API_ADDRESS" env-required:"true"`
	InfluxdbTags    map[string]string `env:"INFLUXDB_TAGS" env-default:""`
	RabbitUrl       RabbitUrlType     `env:"RABBIT_URL" env-default:""`
	RabbitExchange  string            `env:"RABBIT_EXCHANGE" env-default:"v2fly-usage"`
	CheckpointPath  PathType          `env:"CHECKPOINT_PATH" env-default:"./storage/checkpoints"`
	UpdateInterval  int               `env:"UPDATE_INTERVAL" env-default:"5"`
	LogLevel        LogLevelType      `env:"LOG_LEVEL" env-default:"warning"`
}

func (f *PathType) AsString() string {
	return string(*f)
}
func (f *RabbitUrlType) AsString() string {
	return string(*f)
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
func (f *RabbitUrlType) SetValue(s string) error {
	if s == "" {
		log.Warning("RABBIT_URL not defined")
	} else {
		e := validateUrl(&s, []string{"amqp"}, true, true, true, true)
		if e != nil {
			return e
		}
	}
	*f = RabbitUrlType(s)
	return nil
}
func (f *InfluxUrlType) SetValue(s string) error {
	e := validateUrl(&s, []string{"http", "https"}, true, true, true, true)
	if e != nil {
		return e
	}
	*f = InfluxUrlType(s)
	return nil
}
func (f *V2rayUrlType) SetValue(s string) error {
	e := validateUrl(&s, []string{"grpc"}, false, false, true, false)
	if e != nil {
		return e
	}
	*f = V2rayUrlType(s)
	return nil
}

func validateUrl(s *string, scheme_req []string, user_req bool, passwd_req bool, port_req bool, path_req bool) error {
	LogWithRaw := log.WithField("value", s)
	u, err := url.Parse(*s)
	if err != nil {
		LogWithRaw.WithError(err).Error("error while parsing config")
		return err
	}
	_scheme_ok := false
	for _, s := range scheme_req {
		if u.Scheme == s {
			_scheme_ok = true
		}
	}
	if !_scheme_ok {
		e := errors.New("scheme is not acceptable")
		LogWithRaw.WithField("accepts", scheme_req).Error(e)
		return e
	}
	if u.Host == "" {
		e := errors.New("host not provided")
		LogWithRaw.Error(e)
		return e
	}
	if u.User.Username() == "" && user_req {
		e := errors.New("user not provided")
		LogWithRaw.Error(e)
		return e
	}
	token, _ := u.User.Password()
	if token == "" && passwd_req {
		e := errors.New("password not provided")
		LogWithRaw.Error(e)
		return e
	}
	if u.Port() == "" && port_req {
		e := errors.New("port not provided")
		LogWithRaw.Error(e)
		return e
	}
	if strings.Trim(u.Path, "/") == "" && path_req {
		e := errors.New("path not provided")
		LogWithRaw.Error(e)
		return e
	}
	return nil
}
