package supervisor

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anonopiran/Fly2Stats/internal/config"
	"github.com/anonopiran/Fly2Stats/internal/db"
	"github.com/anonopiran/Fly2Stats/internal/notify"
	"github.com/anonopiran/Fly2Stats/internal/v2ray"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

const LOCK_FILE_NAME = "lock"

type Supervisor struct {
	DownServerDB *gorm.DB
	UpSrvs       map[uint]*v2ray.UpServer
	Inflx        db.InfluxCfg
	Rbt          notify.RbtConfig
	FilePath     string
	runInterval  uint
}

// ...Lock Manager
func (sup *Supervisor) isLocked() bool {
	if _, err := os.Stat(filepath.Join(sup.FilePath, LOCK_FILE_NAME)); err == nil || !os.IsNotExist(err) {
		return true
	}
	return false
}
func (sup *Supervisor) Lock(errMsg error) error {
	file, err := os.Create(filepath.Join(sup.FilePath, LOCK_FILE_NAME))
	if err != nil {
		return fmt.Errorf("error creating lock file: %s", err)
	}
	defer file.Close()
	file.WriteString(errMsg.Error())
	return nil
}

// ...down server manager
func (sup *Supervisor) addDownServer(upId uint, ipAddr string) error {
	dbRes := sup.DownServerDB.Create(newDownServer(upId, ipAddr))
	if dbRes.Error != nil {
		return fmt.Errorf("error adding downserver to db: %s", dbRes.Error)
	}
	return nil
}
func (sup *Supervisor) rmDownServer(ipAddr string) error {
	dbRes := sup.DownServerDB.Where("ip_address = ?", ipAddr).Delete(&DownServerType{})
	if dbRes.Error != nil {
		return fmt.Errorf("error deleting down server: %s", dbRes.Error)
	}
	return nil
}
func (sup *Supervisor) addDownServerMany(ipAddrs *mapset.Set[string], upSrvId uint, logger *logrus.Entry) {
	for dIp := range (*ipAddrs).Iter() {
		ll := logger.WithField("ip", dIp)
		ll.Info("server UP")
		if err := sup.addDownServer(upSrvId, dIp); err != nil {
			ll.WithError(err).Error("error adding new down server")
		}
	}
}
func (sup *Supervisor) rmDownServerMany(ipAddrs *mapset.Set[string], logger *logrus.Entry) {
	for dIp := range (*ipAddrs).Iter() {
		ll := logger.WithField("ip", dIp)
		ll.Info("server DOWN")
		if err := sup.rmDownServer(dIp); err != nil {
			ll.WithError(err).Error("error removing down server")
		}
	}
}
func (sup *Supervisor) getAllDownServers() ([]DownServerType, error) {
	dnSrvs := []DownServerType{}
	if err := sup.DownServerDB.Find(&dnSrvs).Error; err != nil {
		return nil, err
	}
	return dnSrvs, nil
}
func (sup *Supervisor) GetDownServerIps(upId uint) (mapset.Set[string], error) {
	downSrvList := []DownServerType{}
	err := sup.DownServerDB.Where("up_srv_id = ?", upId).Select("ip_address").Find(&downSrvList).Error
	if err != nil {
		return nil, fmt.Errorf("error getting down servers from db %s", err)
	}
	ipSet := mapset.NewSet[string]()
	for _, dSrv := range downSrvList {
		ipSet.Add(dSrv.IpAddress)
	}
	return ipSet, nil
}
func (sup *Supervisor) conn(downSrv *DownServerType) (*grpc.ClientConn, error) {
	sup.logger(downSrv).Debug("dialing")
	return v2ray.NewInsecureGrpc(net.ParseIP(downSrv.IpAddress), sup.UpSrvs[downSrv.UpSrvId].Address.Port())
}

// ... query
func (sup *Supervisor) Query() error {
	dsList, err := sup.getAllDownServers()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	ctx := context.Background()
	ctx, cl := context.WithTimeout(ctx, 5*time.Second)
	defer cl()
	if sup.isLocked() {
		return fmt.Errorf("lock file exists")
	}
	logrus.Debugf("servers to query: %+v", dsList)
	for _, dSrv := range dsList {
		dSrv := dSrv
		wg.Add(1)
		go func(_dSrv *DownServerType) {
			defer wg.Done()
			ll := sup.logger(_dSrv)
			_uSrv := sup.UpSrvs[_dSrv.UpSrvId]
			conn, err := sup.conn(_dSrv)
			if err != nil {
				ll.WithError(err).Error("can not dial down server")
				return
			}
			ll.Debug("quering")
			_dSrvStatList, err := _uSrv.GetStats(ctx, conn)
			if err != nil {
				ll.WithError(err).Error("can not query down server stats. locking ...")
				if _, ok := err.(v2ray.DesructiveQueryErr); ok {
					if err := sup.Lock(err); err != nil {
						ll.WithError(err).Error("can not lock!")
					}
				}
				return
			}
			for _, _dSrvStat := range *_dSrvStatList {
				st := db.StatRecordType{UserStatType: _dSrvStat, ServerUri: _uSrv.Address.String(), ServerIp: _dSrv.IpAddress}
				err := st.ToFile(sup.FilePath)
				if err != nil {
					ll.WithError(err).Error("can not write checkpoint file. locking ...")
					if err := sup.Lock(err); err != nil {
						ll.WithError(err).Error("can not lock!")
					}
				}
			}
		}(&dSrv)
	}
	wg.Wait()
	return nil
}

// ...
func (sup *Supervisor) logger(downSrv *DownServerType) *logrus.Entry {
	return sup.UpSrvs[downSrv.UpSrvId].Logger(downSrv.logger(nil))
}

// ...
func (sup *Supervisor) Start() {
	sleeper := time.NewTicker(time.Second * time.Duration(sup.runInterval))
	defer sleeper.Stop()
	for {
		sup.ServiceDiscovery()
		// sup.logger(nil).WithField("data", sup).Debug("current upservers")
		sup.RunOnce()
		<-sleeper.C
	}
}

// ...
func (sup *Supervisor) ServiceDiscovery() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	for upSrvId, upSrv := range sup.UpSrvs {
		wg.Add(1)
		upSrvId := upSrvId
		upSrv := upSrv
		go func(_upSrv *v2ray.UpServer, _upSrvId uint) {
			defer wg.Done()
			ll := _upSrv.Logger(nil)
			newDownServers := _upSrv.Discover(ctx)
			ll.WithField("ips", newDownServers).Debug("discovered")
			currDownServers, err := sup.GetDownServerIps(_upSrvId)
			ll.WithField("ips", currDownServers).Debug("current")
			if err != nil {
				ll.WithError(err).Error("error getting down servers")
				return
			}
			downIp := currDownServers.Difference(newDownServers)
			upIp := newDownServers.Difference(currDownServers)
			sup.rmDownServerMany(&downIp, ll)
			sup.addDownServerMany(&upIp, _upSrvId, ll)
		}(upSrv, upSrvId)
	}
	wg.Wait()
}
func (sup *Supervisor) RunOnce() {
	ctx := context.Background()
	err := sup.Query()
	if err != nil {
		logrus.WithError(err).Error("can not query stats", err)
		return
	}
	updates := db.SyncDBRecords(&sup.Inflx, ctx, sup.FilePath)
	notify.Notify(&sup.Rbt, updates)
}

// ...
func NewSupervisor(cfg config.SupervisorConfigType, upstreamCfg config.UpstreamConfigType) (*Supervisor, error) {
	downServerdb, err := newDownServerDB()
	if err != nil {
		return nil, fmt.Errorf("error creating supervisor (down server db): %s", err)
	}
	upServList := map[uint]*v2ray.UpServer{}
	for c, u := range upstreamCfg.Address {
		upServList[uint(c)], err = v2ray.NewServer(u)
		if err != nil {
			return nil, fmt.Errorf("can not create upstream server: %s", err)
		}
	}
	sup := Supervisor{
		runInterval:  cfg.Interval,
		UpSrvs:       upServList,
		DownServerDB: downServerdb,
		Inflx: db.InfluxCfg{
			Url:    cfg.InfluxdbUrl.GetURL(),
			Token:  cfg.InfluxdbUrl.GetToken(),
			Org:    cfg.InfluxdbUrl.GetOrg(),
			Bucket: cfg.InfluxdbUrl.GetBucket(),
			Tags:   cfg.InfluxdbTags,
		},
		Rbt: notify.RbtConfig{
			URL:          cfg.RabbitUrl.String(),
			ExchangeName: cfg.RabbitExchange,
		},
		FilePath: cfg.CheckpointPath,
	}
	return &sup, nil
}
