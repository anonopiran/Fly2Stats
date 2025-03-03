package v2ray

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/anonopiran/Fly2Stats/internal/config"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IServer interface {
	NewStatRequest() (IStatsRequest, error)
	ParseStats([]IStat) ([]UserStatType, error)
	NewServiceClient(conn *grpc.ClientConn) StatsServiceClient
}
type UpServer struct {
	IServer IServer
	Address config.UpstreamUrlType
}

// ...
func (v *UpServer) Discover(ctx context.Context) mapset.Set[string] {
	ll := v.Logger(nil)
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", v.Address.Hostname())
	if err != nil {
		errLL := logrus.WithField("server", v.Address.Hostname()).WithError(err)
		errMsg := "error looking up server"
		if strings.HasSuffix(err.Error(), "no such host") {
			errLL.Warn(errMsg)
		} else {
			errLL.Error(errMsg)
		}
		ips = []net.IP{}
	} else {
		ll.Debugf("found ips %+v", ips)
	}
	ipSet := mapset.NewSet[string]()
	for _, ip := range ips {
		ipSet.Add(ip.String())
	}
	return ipSet
}
func (v *UpServer) GetStats(ctx context.Context, conn *grpc.ClientConn) (*[]UserStatType, error) {
	if conn == nil {
		return nil, ErrNilConnection
	}
	serviceClient := v.IServer.NewServiceClient(conn)
	req, err := v.IServer.NewStatRequest()
	if err != nil {
		return nil, fmt.Errorf("error creating StatRequest: %s", err)
	}
	res, err := serviceClient.QueryStats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error query stats: %s", err)
	}
	data, err := v.IServer.ParseStats(res)
	if err != nil {
		return nil, DesructiveQueryErr{fmt.Errorf("error parsing stats: %s", err)}
	}
	return &data, nil
}
func (v *UpServer) Logger(ll *logrus.Entry) *logrus.Entry {
	if ll == nil {
		ll = logrus.NewEntry(logrus.StandardLogger())
	}
	return ll.WithField("upstream", v.Address.String())
}

// ...
func NewServer(srv config.UpstreamUrlType) (*UpServer, error) {
	var upSrv *UpServer
	switch srv.ServerType {
	case config.V2FLY_SRV:
		upSrv = NewV2flyServer(srv)
	case config.XRAY_SRV:
		upSrv = NewXrayServer(srv)
	default:
		return nil, ErrUnknownServerType(srv)
	}
	return upSrv, nil
}
