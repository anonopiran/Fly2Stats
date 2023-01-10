package stats2influx

import (
	config "Fly2Stats/Config"
	u2s "Fly2Stats/Upstream2Stats"
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func Write(points u2s.UserStatListTypes) ([]string, error) {
	cfg := config.Config()
	server := cfg.InfluxdbUrl.AsUrl()
	token, _ := server.User.Password()
	client := influxdb2.NewClient(server.Scheme+"://"+server.Host, token)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(server.User.Username(), strings.Trim(server.Path, "/"))
	writeAPI.EnableBatching()
	updates := []string{}
	for _, v_ := range points {
		tags := map[string]string{"user": v_.Username, "direction": string(v_.Direction), "server_url": v_.ServerUri, "server_ip": v_.ServerIp}
		for k_, v_ := range cfg.InfluxdbTags {
			tags[k_] = v_
		}
		p := influxdb2.NewPoint("bandwidth",
			tags, map[string]interface{}{"used": v_.Value}, time.Unix(v_.Time, 0))
		log.WithField("data", fmt.Sprintf("%+v", p)).Debug("writing to influx")
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			log.WithError(err).Error("could not write to influx db after %d writes\n", len(updates))
			return updates, err
		}
		updates = append(updates, v_.Username)
	}
	err := writeAPI.Flush(context.Background())
	if err != nil {
		log.WithError(err).Errorf("could not write to influx db after %d writes\n", len(updates))
		return updates, err
	}
	return updates, nil
}
