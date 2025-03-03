package config

import (
	"errors"
	"net/url"
)

type UpstreamUrlType struct {
	url.URL
	ServerType ServerTypeEnumType
}

type UpstreamConfigType struct {
	Address []UpstreamUrlType `koanf:"address" validate:"required,min=1"`
}

// ...
type ServerTypeEnumType string

const (
	V2FLY_SRV ServerTypeEnumType = "v2fly"
	XRAY_SRV  ServerTypeEnumType = "xray"
)

// ...
func (f *UpstreamUrlType) UnmarshalText(text []byte) error {
	u, err := validateUrl(string(text), []string{"grpc"}, true, false, true, false)
	if err != nil {
		return err
	}
	var srvType ServerTypeEnumType
	srvTypeConfig := u.User.Username()
	switch ss := ServerTypeEnumType(srvTypeConfig); ss {
	case V2FLY_SRV, XRAY_SRV:
		srvType = ss
	default:
		return errors.New("servertype not understood")
	}
	u.User = nil
	*f = UpstreamUrlType{URL: *u, ServerType: srvType}
	return nil
}
