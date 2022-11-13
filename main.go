package main

import (
	config "Fly2Stats/Config"
	grpc2stats "Fly2Stats/Grpc2Stats"
	stats2influx "Fly2Stats/Stats2Influx"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

var cfg config.SettingsType
var influx_server stats2influx.InfluxServer

func stat_to_file() (string, error) {
	file_name := filepath.Join(cfg.Checkpoint_Path, fmt.Sprintf("%v.json", time.Now().Unix()))
	f, err := os.Create(file_name)
	if err != nil {
		log.WithError(err).Error("can not open file")
		return "", err
	}
	defer f.Close()
	stats, err := grpc2stats.ReadStats(cfg.V2fly_Api_Address, false)
	if err != nil {
		return "", err
	}
	jstats, err := json.Marshal(stats)
	if err != nil {
		log.WithField("stats", fmt.Sprintf("%v+", stats)).WithError(err).Error("error converting stats to json")
		return "", err
	}
	_, err = f.Write(jstats)
	if err != nil {
		log.WithField("json stats", string(jstats)).WithError(err).Error("error saving data")
		return "", err
	}
	grpc2stats.ReadStats(cfg.V2fly_Api_Address, true)
	return file_name, nil
}
func files_to_influx() (int, error) {
	files, err := os.ReadDir(cfg.Checkpoint_Path)
	if err != nil {
		log.WithError(err).Error("can not read checkpoint files")
		return 0, err
	}
	cnt_total := 0
	for _, f_ := range files {
		f_path := filepath.Join(cfg.Checkpoint_Path, f_.Name())
		f_file, err := os.ReadFile(f_path)
		if err != nil {
			log.WithError(err).Error("can not read json file")
			continue
		}
		f_data := make(grpc2stats.Stats)
		err = json.Unmarshal(f_file, &f_data)
		if err != nil {
			log.WithError(err).Error("can not parse json file")
			continue
		}
		cnt, wr_er := stats2influx.Write(influx_server, f_data)
		cnt_total += cnt
		if wr_er {
			log.Debug("some errors occured while writing to influx. keeping checkpoint file ...")
			continue
		}
		log.WithField("file", f_path).Debug("removing file")
		err = os.Remove(f_path)
		if err != nil {
			log.WithError(err).Error("can not delete checkpoint file")
		}
	}

	return cnt_total, nil
}
func run() {
	log.Error("test err")
	cfg = config.Config()
	influx_server = stats2influx.InfluxServer{InfluxURI: cfg.Influxdb_Url, InfluxOrg: cfg.Influxdb_Org, InfluxToken: cfg.Influxdb_Token, InfluxBucket: cfg.Influxdb_Bucket}
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.Interval))
	for {
		_, err := stat_to_file()
		if err != nil {

		} else {
			count, _ := files_to_influx()
			f_ := log.Fields{"sleep": cfg.Interval, "count": count}
			log.WithFields(f_).Info("stats saved to influx")
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
