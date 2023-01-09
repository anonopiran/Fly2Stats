package upstream2stats

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DirectionType string

const (
	Downlink DirectionType = "downlink"
	Uplink   DirectionType = "uplink"
)

type UserStatType struct {
	Username  string
	Time      int64
	Direction DirectionType
	Value     int64
	ServerUri string
	ServerIp  string
}

type UserStatListTypes []UserStatType

type ServerType struct {
	Uri  string
	Ip   net.IP
	Port int64
}

func (s *UserStatListTypes) SaveJson(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	jstats, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = f.Write(jstats)
	if err != nil {
		return err
	}
	return nil
}

func (srv *ServerType) DialGrpc() (*grpc.ClientConn, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", srv.Ip.String(), srv.Port), opt)
	return conn, err
}
