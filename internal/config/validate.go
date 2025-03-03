package config

import (
	"fmt"
	"net/url"
	"strings"
)

func validateUrl(s string, scheme_req []string, user_req bool, passwd_req bool, port_req bool, path_req bool) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("can not parse url (%s): %s", s, err)
	}
	_scheme_ok := false
	for _, s := range scheme_req {
		if u.Scheme == s {
			_scheme_ok = true
		}
	}
	if !_scheme_ok {
		return nil, fmt.Errorf("scheme should be one of %+v (%s)", scheme_req, s)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("hostname should be provided (%s)", s)
	}
	if u.User.Username() == "" && user_req {
		return nil, fmt.Errorf("username should be provided (%s)", s)
	}
	token, _ := u.User.Password()
	if token == "" && passwd_req {
		return nil, fmt.Errorf("password should be provided (%s)", s)
	}
	if u.Port() == "" && port_req {
		return nil, fmt.Errorf("port should be provided (%s)", s)
	}
	if strings.Trim(u.Path, "/") == "" && path_req {
		return nil, fmt.Errorf("path should be provided (%s)", s)
	}
	return u, nil
}
