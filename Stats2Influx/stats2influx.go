package stats2influx

import (
	grpc2stats "Fly2Stats/Grpc2Stats"
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// ==================================
// types
// ==================================
type InfluxServer struct {
	InfluxURI    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string
	InfluxTags   map[string]string
}

// ==================================
// functions
// ==================================
func Write(server InfluxServer, points grpc2stats.UserStatListTypes) ([]string, error) {
	client := influxdb2.NewClient(server.InfluxURI, server.InfluxToken)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(server.InfluxOrg, server.InfluxBucket)
	writeAPI.EnableBatching()
	updates := []string{}
	for _, v_ := range points {
		tags := map[string]string{"user": v_.Username, "direction": string(v_.Direction), "server_url": v_.ServerUri, "server_ip": v_.ServerIp}
		for k_, v_ := range server.InfluxTags {
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
