package v2ray

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ...
type GrpcError struct {
	Err error
}

func (err *GrpcError) Error() string {
	return err.Err.Error()
}

// ...
func NewInsecureGrpc(address net.IP, port string) (*grpc.ClientConn, error) {
	if address == nil {
		return nil, ErrNilAddress
	}
	if port == "" {
		return nil, ErrNilPort
	}
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", address.String(), port), opt)
	if err != nil {
		return nil, fmt.Errorf("error dial %s:%s:%s", address.String(), port, err)
	}
	return conn, nil
}
