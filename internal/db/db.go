package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
)

func LoadAllRecords(dir string) ([]*StatRecordType, error) {
	filesList, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("can not list checkpoint files")
	}
	allRecords := []*StatRecordType{}
	for _, f_ := range filesList {
		ll := logrus.WithField("file", f_)
		if !(strings.HasSuffix(f_.Name(), ".json") && strings.HasPrefix(f_.Name(), "checkpoint-")) {
			continue
		}
		f_path := filepath.Join(dir, f_.Name())
		f_file, err := os.ReadFile(f_path)
		if err != nil {
			ll.WithError(err).Error("can not read checkpoint file. skip ...")
			continue
		}
		f_data := StatRecordType{}
		err = json.Unmarshal(f_file, &f_data)
		if err != nil {
			ll.WithError(err).Error("can not parse checkpoint file. skip ...")
			continue
		}
		allRecords = append(allRecords, &f_data)
	}
	return allRecords, nil
}
func SyncDBRecords(influxCfg *InfluxCfg, ctx context.Context, checkPointDir string) []string {
	client := influxdb2.NewClient(influxCfg.Url, influxCfg.Token)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(influxCfg.Org, influxCfg.Bucket)
	updates := []*StatRecordType{}
	points, err := LoadAllRecords(checkPointDir)
	if err != nil {
		logrus.WithError(err).Error("can not load checkpoint files")
		return nil
	}
	for _, v_ := range points {
		p := v_.AsInflux(influxCfg.Tags)
		logrus.WithField("data", fmt.Sprintf("%+v", p)).Debug("writing to influx")
		err := writeAPI.WritePoint(ctx, p)
		if err != nil {
			logrus.WithField("file", v_.file).WithError(err).Error("could not write to influx db. skip ...")
			continue
		}
		updates = append(updates, v_)
	}
	err = writeAPI.Flush(ctx)
	if err != nil {
		logrus.WithError(err).Error("could not flush to influx db.")
	}
	result := []string{}
	for _, u := range updates {
		if err := u.DelFile(); err != nil {
			logrus.WithField("file", u.file).WithError(err).Error("can not delete record file")
		}
		result = append(result, u.Username)
	}
	return result
}
