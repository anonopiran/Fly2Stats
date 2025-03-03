package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anonopiran/Fly2Stats/internal/v2ray"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2write "github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxCfg struct {
	Url    string
	Token  string
	Org    string
	Bucket string
	Tags   map[string]string
}
type StatRecordType struct {
	v2ray.UserStatType
	ServerUri string
	ServerIp  string
	file      string
}

func (s *StatRecordType) AsInflux(ExtraTags map[string]string) *influxdb2write.Point {
	tags := map[string]string{"user": s.Username, "direction": string(s.Direction), "server_url": s.ServerUri, "server_ip": s.ServerIp}
	for k_, v_ := range ExtraTags {
		tags[k_] = v_
	}
	p := influxdb2.NewPoint("bandwidth", tags, map[string]interface{}{"used": s.Value}, time.Unix(s.Time, 0))
	return p
}
func (s *StatRecordType) ToFile(dir string) error {
	name := fmt.Sprintf("checkpoint-%s-%s-%d-%s-%s.json", s.Username, s.Direction, s.Time, s.ServerUri, s.ServerIp)
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("can not create checkpoint file: %s", err)
	}
	defer f.Close()
	jstats, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("can not marshall checkpoint file: %s", err)
	}
	_, err = f.Write(jstats)
	if err != nil {
		return fmt.Errorf("can not write checkpoint file: %s", err)
	}
	return nil
}
func (s *StatRecordType) DelFile() error {
	if s.file != "" {
		return os.Remove(s.file)
	} else {
		return fmt.Errorf("StatRecordType doesn't have any file")
	}
}
