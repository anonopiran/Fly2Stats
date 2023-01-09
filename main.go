package main

import (
	config "Fly2Stats/Config"
	faillock "Fly2Stats/FailLock"
	stats2influx "Fly2Stats/Stats2Influx"
	"fmt"

	u2s "Fly2Stats/Upstream2Stats"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"
)

var cfg config.SettingsType
var redisClient redis.Client
var redisActivated bool

func Stats2File() {
	faillock.TestUp()
	for _, trg := range config.Config.V2flyApiAddress {
		logWithTarget := log.WithField("target", trg)
		res, err := u2s.ResolveV2FlyServer(trg.AsUrl())
		if err != nil {
			logWithTarget.WithError(err).Error("error while resolving v2fly server")
			continue
		}
		for _, srv := range res {
			logWithServer := logWithTarget.WithField("server", srv)
			stats, err := u2s.ReadUpstream(&srv, true)
			if err != nil {
				logWithServer.WithError(err).Error("error while reading stats")
				continue
			}
			if len(stats) == 0 {
				logWithServer.Debug("nothing to write")
				continue
			}
			filePath := filepath.Join(cfg.CheckpointPath.AsString(), fmt.Sprintf("checkpoint-%v-%s-%s.json", time.Now().Unix(), srv.Uri, srv.Ip))
			err = stats.SaveJson(filePath)
			if err != nil {
				logWithServer.WithField("stats", stats).Error("fatal error while saving stats")
				faillock.LockAndPan(err)
			}
			logWithServer.WithField("count", len(stats)).Info("stats saved to file")
		}
	}
}
func Files2Influx() []string {
	updates_total := []string{}
	files, err := os.ReadDir(cfg.CheckpointPath.AsString())
	if err != nil {
		log.WithError(err).Errorf("can not read checkpoint files at %n", cfg.CheckpointPath)
		return updates_total
	}
	for _, f_ := range files {
		if !strings.HasSuffix(f_.Name(), ".json") {
			continue
		}
		f_path := filepath.Join(cfg.CheckpointPath.AsString(), f_.Name())
		f_file, err := os.ReadFile(f_path)
		if err != nil {
			log.WithError(err).Errorf("can not read json file at %s\n", f_path)
			continue
		}
		f_data := *new(u2s.UserStatListTypes)
		err = json.Unmarshal(f_file, &f_data)
		if err != nil {
			log.WithError(err).Error("can not parse json file at %s\n", f_path)
			continue
		}
		updates, err := stats2influx.Write(f_data)
		if err != nil {
			log.WithError(err).Error("errors occured while writing to influx. keeping checkpoint file ...")
			continue
		}
		updates_total = append(updates_total, updates...)
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
	if len(updates_total_unique) > 0 {
		log.WithField("count", len(updates_total_unique)).Info("stats saved to influx")
	}
	return updates_total_unique
}
func Notify(updates []string) error {
	if redisActivated && len(updates) > 0 {
		update_json, _ := json.Marshal(updates)
		if err := redisClient.Publish(context.Background(), "v2fly-update-stats", update_json).Err(); err != nil {
			log.WithError(err).WithField("data", update_json).Error("error while publishing state update event")
			return err
		}
	}
	return nil
}
func run() {
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.UpdateInterval))
	for {
		Stats2File()
		updates := Files2Influx()
		Notify(updates)
		<-sleeper.C
	}
}
func help() {
	config.Describe()
}
func init() {
	cfg = config.Config
	redisActivated = (cfg.RedisUrl != "")
	if redisActivated {
		rdsOpts := cfg.RedisUrl.AsOpts()
		redisClient = *redis.NewClient(&rdsOpts)
	}
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
