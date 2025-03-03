package config

import (
	"fmt"
	"net/url"
	"strings"
)

type InfluxUrlType struct{ url.URL }
type RabbitUrlType struct{ url.URL }
type TagMapType map[string]string
type SupervisorConfigType struct {
	Interval       uint          `koanf:"interval" validate:"required"`
	InfluxdbUrl    InfluxUrlType `koanf:"influxdb_url" validate:"required"`
	InfluxdbTags   TagMapType    `koanf:"influxdb_tags"`
	RabbitUrl      RabbitUrlType `koanf:"rabbit_url"`
	RabbitExchange string        `koanf:"rabbit_exchange" validate:"required_with=RabbitUrl"`
	CheckpointPath string        `koanf:"checkpoint_path" validate:"dirpath"`
}

// ...
func (f *RabbitUrlType) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	u, err := validateUrl(s, []string{"amqp"}, true, true, true, true)
	if err != nil {
		return err
	}
	*f = RabbitUrlType{*u}
	return nil
}
func (f *InfluxUrlType) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	u, err := validateUrl(string(text), []string{"http", "https"}, true, true, true, true)
	if err != nil {
		return err
	}
	*f = InfluxUrlType{*u}
	return nil
}
func (f *TagMapType) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			result[key] = value
		} else {
			return fmt.Errorf("env variable for map item not understood %s", pair)
		}
	}
	*f = result
	return nil
}

// ...
func (f *InfluxUrlType) GetURL() string {
	return f.Scheme + "://" + f.Host
}
func (f *InfluxUrlType) GetToken() string {
	p, _ := f.User.Password()
	return p
}
func (f *InfluxUrlType) GetOrg() string {
	return f.User.Username()
}
func (f *InfluxUrlType) GetBucket() string {
	return strings.Trim(f.Path, "/")
}
