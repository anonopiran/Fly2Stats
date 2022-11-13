package grpc2stats

import (
	"context"
	"fmt"
	"time"

	"regexp"

	"github.com/v2fly/v2ray-core/v5/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"
)

// ==================================
// types
// ==================================
type UserStat struct {
	Downlink int64
	Uplink   int64
	Time     int64
}
type Stats map[string]UserStat

// ==================================
// functions
// ==================================
func query_stats(server string, reset bool) (*command.QueryStatsResponse, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(server, opt)
	if err != nil {
		log.WithError(err).Error("fail to dial grpc server")
		return nil, err
	}
	defer conn.Close()
	client := command.NewStatsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := command.QueryStatsRequest{Reset_: reset, Regexp: true, Pattern: "user.+"}
	r, err := client.QueryStats(ctx, &req)
	if err != nil {
		log.WithError(err).Error("could not query stats")
		return nil, err
	}
	log.WithField("data", r).Debugln("read v2fly stats")
	return r, nil
}
func parse_stats(r *command.QueryStatsResponse) Stats {
	ct := time.Now().Unix()
	stats := make(Stats)
	reg, _ := regexp.Compile("^.+>>>(.+?)>>>.+>>>(.+)$")
	for _, s_ := range r.Stat {
		reg_res := reg.FindStringSubmatch(s_.Name)
		u_ := reg_res[1]
		t_ := reg_res[2]
		entry, ok := stats[u_]
		if !ok {
			entry = UserStat{Time: ct}
		}
		if t_ == "downlink" {
			entry.Downlink = s_.Value
		} else if t_ == "uplink" {
			entry.Uplink = s_.Value
		}
		stats[u_] = entry
	}
	log.WithField("data", fmt.Sprintf("%+v", stats)).Debug("parse user stat")
	return stats
}
func ReadStats(server string, reset bool) (Stats, error) {
	r, err := query_stats(server, reset)
	if err != nil {
		return nil, err
	}
	stats := parse_stats(r)
	return stats, nil
}
