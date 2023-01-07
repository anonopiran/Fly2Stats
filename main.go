package main

import (
	config "Fly2Stats/Config"
	grpc2stats "Fly2Stats/Grpc2Stats"
	stats2influx "Fly2Stats/Stats2Influx"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var cfg config.SettingsType
var influx_server stats2influx.InfluxServer
var panic_file_name string

func panic_file(err error) error {
	f, _err := os.Create(panic_file_name)
	if _err != nil {
		log.WithError(_err).Errorf("can not open panic file at %s", panic_file_name)
		return _err
	}
	defer f.Close()
	f.WriteString(err.Error())
	return nil
}
func stat_to_file() (string, error) {
	if _, err := os.Stat(panic_file_name); err == nil {
		log.Panicf("panic file exists at %s", panic_file_name)
	}
	err := panic_file(errors.New("test"))
	if err != nil {
		log.WithError(err).Panicf("can not open test panic file at %s", panic_file_name)
	}
	err = os.Remove(panic_file_name)
	if err != nil {
		log.WithError(err).Panicf("can not remove test panic file at %s", panic_file_name)
	}
	stats, err := grpc2stats.ReadStats(cfg.V2fly_Api_Address, true)
	if err != nil {
		log.WithError(err).Error("error reading stats from v2fly")
		return "", err
	}
	if len(stats) == 0 {
		log.Info("nothing to write")
		return "", nil
	}
	file_name := filepath.Join(cfg.Checkpoint_Path, fmt.Sprintf("checkpoint-%v.json", time.Now().Unix()))
	f, err := os.Create(file_name)
	if err != nil {
		panic_file(err)
		log.WithError(err).Panicf("can not open the file %s", file_name)
	}
	defer f.Close()
	jstats, err := json.Marshal(stats)
	if err != nil {
		panic_file(err)
		log.WithField("stats", fmt.Sprintf("%v+", stats)).WithError(err).Panic("error converting stats to json")
	}
	_, err = f.Write(jstats)
	if err != nil {
		panic_file(err)
		log.WithField("json stats", string(jstats)).WithError(err).Panicf("error saving data to %s", file_name)
	}
	return file_name, nil
}
func files_to_influx() (int, error) {
	files, err := os.ReadDir(cfg.Checkpoint_Path)
	if err != nil {
		log.WithError(err).Errorf("can not read checkpoint files at %n", cfg.Checkpoint_Path)
		return 0, err
	}
	cnt_total := 0
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
		f_data := *new(grpc2stats.Stats)
		err = json.Unmarshal(f_file, &f_data)
		if err != nil {
			log.WithError(err).Error("can not parse json file at %s\n", f_path)
			continue
		}
		cnt, err := stats2influx.Write(influx_server, f_data)
		cnt_total += cnt
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
	return cnt_total, nil
}
func run() {
	cfg = config.Config()
	panic_file_name = filepath.Join(cfg.Checkpoint_Path, "panic")
	influx_server = stats2influx.InfluxServer{InfluxURI: cfg.Influxdb_Url, InfluxOrg: cfg.Influxdb_Org, InfluxToken: cfg.Influxdb_Token, InfluxBucket: cfg.Influxdb_Bucket, InfluxTags: cfg.Influxdb_Tags}
	sleeper := time.NewTicker(time.Second * time.Duration(cfg.Update_Interval))
	for {
		_, err := stat_to_file()
		if err != nil {
		} else {
			count, _ := files_to_influx()
			f_ := log.Fields{"sleep": cfg.Update_Interval, "count": count}
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
