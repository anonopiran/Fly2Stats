package main

import (
	config "Fly2Stats/Config"
	grpc2stats "Fly2Stats/Grpc2Stats"
	stats2influx "Fly2Stats/Stats2Influx"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"
)

var cfg = config.Config()
var influx_server = stats2influx.InfluxServer{InfluxURI: cfg.Influxdb_Url, InfluxOrg: cfg.Influxdb_Org, InfluxToken: cfg.Influxdb_Token, InfluxBucket: cfg.Influxdb_Bucket, InfluxTags: cfg.Influxdb_Tags}
var panic_file_name = filepath.Join(cfg.Checkpoint_Path, "panic")
var redis_available = false
var redis_client redis.Client

func redis_setup() {
	if cfg.Redis_Url != "" {
		redis_opt, err := redis.ParseURL(cfg.Redis_Url)
		if err != nil {
			log.WithError(err).WithField("uri", cfg.Redis_Url).Panicf("redis uri is defined but nit understood")
		}
		redis_client = *redis.NewClient(redis_opt)
		redis_available = true
	} else {
		log.Warn("redis is not enabled")
	}

}
func resolve_v2fly() []grpc2stats.ServerType {
	servers := []grpc2stats.ServerType{}
	for _, srv := range cfg.V2fly_Api_Address {
		x := strings.Split(srv, ":")
		host := x[0]
		port, err := strconv.ParseInt(x[1], 10, 0)
		if err != nil {
			log.WithError(err).Panicf("can not parse port %s", srv)
		}
		ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
		if err != nil {
			log.WithError(err).Errorf("can not resolve host %s. skiping...", host)
			continue
		}
		for _, ip_ := range ips {
			servers = append(servers, grpc2stats.ServerType{Ip: ip_, Uri: host, Port: port})
		}
	}
	log.Infof("discovered %d services", len(servers))
	log.Debugf("%v", servers)
	return servers
}
func panic_file(panpan bool, nolog bool, err error) {
	f, _err := os.Create(panic_file_name)
	var e_ func(string, ...interface{})
	if panpan {
		e_ = log.WithError(_err).Panicf
	} else {
		e_ = log.WithError(_err).Errorf
	}
	if _err != nil {
		log.WithError(_err).Panicf("can not open panic file at %s", panic_file_name)
	}
	defer f.Close()
	f.WriteString(err.Error())
	if !nolog {
		e_(err.Error())

	}
}
func stat_to_file() {
	if _, err := os.Stat(panic_file_name); err == nil {
		log.Panicf("panic file exists at %s", panic_file_name)
	}
	panic_file(false, true, errors.New("test"))
	err := os.Remove(panic_file_name)
	if err != nil {
		log.WithError(err).Panicf("can not remove test panic file at %s", panic_file_name)
	}
	for _, srv_ := range resolve_v2fly() {
		stats, err := grpc2stats.ReadStats(srv_, true)
		log_ := log.WithField("server", srv_)
		if err != nil {
			log_.WithError(err).Error("error reading stats from v2fly")
			continue
		}
		if len(stats) == 0 {
			log_.Info("nothing to write")
			continue
		}
		file_name := filepath.Join(cfg.Checkpoint_Path, fmt.Sprintf("checkpoint-%v-%s-%s.json", time.Now().Unix(), srv_.Uri, srv_.Ip))
		f, err := os.Create(file_name)
		if err != nil {
			panic_file(true, false, err)
		}
		defer f.Close()
		jstats, err := json.Marshal(stats)
		if err != nil {
			panic_file(true, false, err)
		}
		_, err = f.Write(jstats)
		if err != nil {
			panic_file(true, false, err)
		}
		log_.Infof("write %d records to file", len(stats))
	}
}
func files_to_influx() ([]string, error) {
	updates_total := []string{}
	files, err := os.ReadDir(cfg.Checkpoint_Path)
	if err != nil {
		log.WithError(err).Errorf("can not read checkpoint files at %n", cfg.Checkpoint_Path)
		return updates_total, err
	}
	for _, f_ := range files {
		if !strings.HasSuffix(f_.Name(), ".json") {
			continue
		}
		f_path := filepath.Join(cfg.Checkpoint_Path, f_.Name())
		f_file, err := os.ReadFile(f_path)
		if err != nil {
			log.WithError(err).Errorf("can not read json file at %s\n", f_path)
			continue
		}
		f_data := *new(grpc2stats.UserStatListTypes)
		err = json.Unmarshal(f_file, &f_data)
		if err != nil {
			log.WithError(err).Error("can not parse json file at %s\n", f_path)
			continue
		}
		updates, err := stats2influx.Write(influx_server, f_data)
		updates_total = append(updates_total, updates...)
		if err != nil {
			log.WithError(err).Error("errors occured while writing to influx. keeping checkpoint file ...")
			continue
		}
		log.WithField("file", f_path).Debug("removing file")
		err = os.Remove(f_path)
		if err != nil {
			log.WithError(err).Error("can not delete checkpoint file at %s\n", f_path)
			continue
		}
	}
	updates_total_keys := make(map[string]bool)
	updates_total_unique := []string{}
	for _, entry := range updates_total {
		if _, value := updates_total_keys[entry]; !value {
			updates_total_keys[entry] = true
			updates_total_unique = append(updates_total_unique, entry)
		}
	}
	return updates_total_unique, nil
}
func run() {
	redis_setup()
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.Update_Interval))
	for {
		stat_to_file()
		updates, _ := files_to_influx()
		f_ := log.Fields{"sleep": cfg.Update_Interval, "count": len(updates)}
		log.WithFields(f_).Info("stats saved to influx")
		if redis_available && len(updates) > 0 {
			update_json, _ := json.Marshal(updates)
			if err := redis_client.Publish(context.Background(), "v2fly-update-stats", update_json).Err(); err != nil {
				log.WithError(err).WithField("data", update_json).Error("error while publishing state update event")
			}
		}
		<-sleeper.C
	}
}
func help() {
	config.Describe()
}
func main() {
	f_help := flag.Bool("env", false, "list available env variables")
	flag.Parse()
	if *f_help {
		help()
	} else {
		run()
	}
}
