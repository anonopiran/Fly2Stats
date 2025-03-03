package v2ray

import (
	"errors"
	"fmt"
)

var ErrNilAddress = errors.New("address cannot be nil")
var ErrNilPort = errors.New("port cannot be empty")
var ErrNilRequest = errors.New("request is nil")
var ErrNilConnection = errors.New("connection is nil")

func ErrUnknownRequestType(req interface{}) error {
	return fmt.Errorf("unknown request type: %T", req)
}
func ErrUnknownServerType(req interface{}) error {
	return fmt.Errorf("unknown server type: %T", req)
}
func ErrUnknownServiceClientType(req interface{}) error {
	return fmt.Errorf("unknown handler client type: %T", req)
}

type DesructiveQueryErr struct{ error }
