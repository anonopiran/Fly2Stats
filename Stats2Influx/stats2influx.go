package stats2influx

import (
	grpc2stats "Fly2Stats/Grpc2Stats"
	"context"
	"fmt"
	"reflect"
	"strings"
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
}

// ==================================
// functions
// ==================================
func Write(server InfluxServer, points grpc2stats.Stats) (int, bool) {
	client := influxdb2.NewClient(server.InfluxURI, server.InfluxToken)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(server.InfluxOrg, server.InfluxBucket)
	error_keep := false
	count := 0
	for u_, v_ := range points {
		v_reflect := reflect.ValueOf(v_)
		typeOfS := v_reflect.Type()
		t_ := time.Unix(v_.Time, 0)
		for i := 0; i < v_reflect.NumField(); i++ {
			f_ := typeOfS.Field(i).Name
			if f_ == "Time" {
				continue
			}
			p := influxdb2.NewPoint("bandwidth",
				map[string]string{"user": u_, "direction": strings.ToLower(f_)}, // lower for backward compatibility
				map[string]interface{}{"used": v_reflect.Field(i).Interface()},
				t_)
			log.WithField("data", fmt.Sprintf("%+v", p)).Debug("writing to influx")
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				log.WithError(err).Error("can not write to influx")
				error_keep = true
			}
			count += 1
		}
	}
	return count, error_keep
}
