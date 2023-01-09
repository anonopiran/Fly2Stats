package upstream2stats

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"time"

	"regexp"

	"github.com/v2fly/v2ray-core/v5/app/stats/command"

	log "github.com/sirupsen/logrus"
)

func ResolveV2FlyServer(url url.URL) ([]ServerType, error) {
	logWithUrl := log.WithField("url", url)
	servers := []ServerType{}
	host := url.Hostname()
	port, err := strconv.ParseInt(url.Port(), 10, 0)
	if err != nil {
		logWithUrl.WithError(err).Error("error parsing port number")
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
	if err != nil {
		logWithUrl.WithError(err).Error("can not resolve host")
		return nil, err
	}
	for _, ip_ := range ips {
		servers = append(servers, ServerType{Ip: ip_, Uri: url.Hostname(), Port: port})
	}
	logWithUrl.Debugf("discovered %d services", len(servers))
	logWithUrl.Debugf("%v", servers)
	return servers, nil
}

func ReadUpstream(server *ServerType, reset bool) (UserStatListTypes, error) {
	conn, err := server.DialGrpc()
	if err != nil {
		log.WithError(err).Error("could not dial server")
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
	log.WithField("data", r).Debug("read v2fly stats")
	// ................
	ct := time.Now().Unix()
	stats := *new(UserStatListTypes)
	reg, _ := regexp.Compile("^.+>>>(.+?)>>>.+>>>(.+)$")
	for _, s_ := range r.Stat {
		usage := s_.Value
		if usage == 0 {
			continue
		}
		reg_res := reg.FindStringSubmatch(s_.Name)
		u_ := reg_res[1]                // user email
		t_ := DirectionType(reg_res[2]) // usage direction
		entry := UserStatType{Time: ct, Username: u_, Direction: t_, Value: usage, ServerUri: server.Uri, ServerIp: server.Ip.String()}
		stats = append(stats, entry)
	}
	log.WithField("data", stats).Debug("parse user stat")
	return stats, nil
}
