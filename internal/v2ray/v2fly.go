package v2ray

import (
	"context"
	"regexp"
	"time"

	"github.com/anonopiran/Fly2Stats/internal/config"
	command "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"google.golang.org/grpc"
)

type V2flyServerType struct {
	UpServer
	HandlerFactory func(grpc.ClientConnInterface) command.StatsServiceClient
}

type v2flyStatsServiceClientAdapter struct {
	client command.StatsServiceClient
}

func (a *v2flyStatsServiceClientAdapter) QueryStats(ctx context.Context, in IStatsRequest, opts ...grpc.CallOption) ([]IStat, error) {
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
func (v *V2flyServerType) NewStatRequest() (IStatsRequest, error) {
	req := command.QueryStatsRequest{Reset_: true, Regexp: true, Pattern: "user.+"}
	return &req, nil
}
func (v *V2flyServerType) ParseStats(istat []IStat) ([]UserStatType, error) {
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
func NewV2flyServer(addr config.UpstreamUrlType) *UpServer {
	return &UpServer{Address: addr, IServer: &V2flyServerType{
		HandlerFactory: command.NewStatsServiceClient,
	}}
}
func (v *V2flyServerType) NewServiceClient(conn *grpc.ClientConn) StatsServiceClient {
	hndlr := v.HandlerFactory(conn)
	return &v2flyStatsServiceClientAdapter{hndlr}
}
