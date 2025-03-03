package v2ray

import (
	"context"
	"regexp"
	"time"

	"github.com/anonopiran/Fly2Stats/internal/config"
	command "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
)

type XrayServerType struct {
	UpServer
	HandlerFactory func(grpc.ClientConnInterface) command.StatsServiceClient
}

type xrayStatsServiceClientAdapter struct {
	client command.StatsServiceClient
}

func (a *xrayStatsServiceClientAdapter) QueryStats(ctx context.Context, in IStatsRequest, opts ...grpc.CallOption) ([]IStat, error) {
	statRequest, ok := in.(*command.QueryStatsRequest)
	if !ok {
		return nil, ErrUnknownRequestType(in)
	}
	stList := []IStat{}
	resp, err := a.client.QueryStats(ctx, statRequest, opts...)
	for _, st := range resp.Stat {
		stList = append(stList, IStat(st))
	}
	return stList, err
}

// ...
func (v *XrayServerType) NewStatRequest() (IStatsRequest, error) {
	req := command.QueryStatsRequest{Reset_: true, Pattern: "user"}
	return &req, nil
}
func (v *XrayServerType) ParseStats(istat []IStat) ([]UserStatType, error) {
	stats := []UserStatType{}
	ct := time.Now().Unix()
	reg, _ := regexp.Compile("^user>>>(.+?)>>>traffic>>>(.+)$")
	for _, s_ := range istat {
		usage := s_.GetValue()
		if usage == 0 {
			continue
		}
		reg_res := reg.FindStringSubmatch(s_.GetName())
		u_ := reg_res[1]                // user email
		t_ := DirectionType(reg_res[2]) // usage direction
		entry := UserStatType{Time: ct, Username: u_, Direction: t_, Value: usage}
		stats = append(stats, entry)
	}
	return stats, nil
}
func NewXrayServer(addr config.UpstreamUrlType) *UpServer {
	return &UpServer{Address: addr, IServer: &XrayServerType{
		HandlerFactory: command.NewStatsServiceClient,
	}}
}
func (v *XrayServerType) NewServiceClient(conn *grpc.ClientConn) StatsServiceClient {
	hndlr := v.HandlerFactory(conn)
	return &xrayStatsServiceClientAdapter{hndlr}
}
