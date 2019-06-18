/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tsdb

import (
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/vxconf"
	influxdb "github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type (
	TinyClient interface {
		Init() error
		Shutdown()
		Close()
		FlushStats()

		Add(tags Tags, values Values) error

		Query(ql string) (data interface{}, err error)
	}

	// Point interface {
	// }

	influxClient struct {
		client influxdb.Client
		bp     influxdb.BatchPoints
		rwl    sync.RWMutex

		exitCh chan struct{}
	}

	// InfluxdbPoint struct {
	// 	Tags   string
	// 	Values string
	// }

	Tags   map[string]string
	Values map[string]interface{}
)

var (
	// TODO uniqueClientInfluxDB 初始化时需要 Singleton 加锁 Once；卸载时需要防止反复shutdown或者错误检测导致反复shutdown
	uniqueClientInfluxDB *influxClient
)

func New() TinyClient {
	if uniqueClientInfluxDB != nil {
		return uniqueClientInfluxDB
	}

	cli := &influxClient{}
	if err := cli.Init(); err != nil {
		log.Errorf("cannot init tsdb client: %v", err)
		return nil
	}
	uniqueClientInfluxDB = cli
	log.Debug("influxClient.New() - init ok. connected.")
	return cli
}

func (s *influxClient) Init() error {
	if s.exitCh == nil {
		log.Debug("influxClient.Init() - init exitCh and go routine.")
		s.exitCh = make(chan struct{})
		go s.flushStatsRoutine()
	}
	return nil
}

func (s *influxClient) Shutdown() {
	// log.Infof("   influxClient shutting down.")
	if s.exitCh != nil {
		s.exitCh <- struct{}{}
		time.Sleep(100 * time.Millisecond)
		s.exitCh = nil
	}
	// log.Infof("   influxClient flushStats.")
	s.FlushStats()
	// log.Infof("   influxClient close.")
	s.Close()
	// log.Infof("   influxClient done.")
}

func (s *influxClient) Close() {
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
}

func (s *influxClient) ensureTsdbClient() influxdb.Client {
	if s.client == nil {
		var err error
		// cmdr.SetDefault("server.tsdb.current.addr", DEFAULT_INFLUXDB_SERVER)
		// cmdr.SetDefault("server.tsdb.current.database", conf.AppName)
		// cmdr.SetDefault("server.tdsb.current.mode", MODE_HTTP) // HTTP or UDP
		// cmdr.SetDefault("server.tdsb.current.flush", "100")    // flush to tsdb server each 100 points
		// cmdr.SetDefault("server.tdsb.current.flushTTL", "15")  // flush TTL
		// cmdr.BindEnv("INFLUX_USER")
		// cmdr.BindEnv("INFLUX_PASS")
		s.client, err = influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:     vxconf.GetStringR("server.tsdb.current.addr", ""),
			Username: vxconf.GetStringR("server.tsdb.current.user", ""),
			Password: vxconf.GetStringR("server.tsdb.current.pass", ""),
		})
		if err != nil {
			log.Errorf("Error creating InfluxDB Client: ", err.Error())
		} else {
			return s.client
		}
	} else {
		return s.client
	}
	return nil
}

func (s *influxClient) flushStatsRoutine() {
	heartbeat := time.Tick(time.Duration(vxconf.GetIntR("server.tsdb.current.flushTTL", 16)) * time.Second)
	log.Infof("      [flushStatsRoutine] Routine run. TTL=%ds", vxconf.GetIntR("server.tsdb.current.flushTTL", 16))
	for {
		select {
		case <-s.exitCh:
			goto RET
		case <-heartbeat:
			// log.Infof("      [flushStatsRoutine] Flush stats points to TSDB ...")
			s.FlushStats()
		}
	}
RET:
	log.Infof("      [flushStatsRoutine] Routine stopped.")
}

func (s *influxClient) FlushStats() {
	s.rwl.Lock()
	defer s.rwl.Unlock()

	if s.bp != nil && len(s.bp.Points()) > 0 {
		c := s.ensureTsdbClient()
		if c == nil {
			return
		}

		log.Infof("Flushing %d points to TSDB server ...", len(s.bp.Points()))
		// if err := c.Write(s.bp); err != nil {
		// 	log.Errorf("[TSDB] Write failed. Error: %v", err)
		// }

		bp := s.bp
		// cmdr.SetDefault("server.tsdb.current.database", conf.AppName)
		dbName := vxconf.GetStringR("server.tsdb.current.database", conf.AppName)
		s.bp, _ = influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
			Database:  dbName,
			Precision: "ns",
		})

		// Write the batch
		go func() {
			if err := c.Write(bp); err != nil {
				log.Errorf("[TSDB] Write failed. db=%s. Error: %v", dbName, err)

				s.rwl.Lock()
				defer s.rwl.Unlock()

				s.bp.AddPoints(bp.Points())
			}
		}()
	}
}

func (s *influxClient) Add(tags Tags, values Values) error {
	// if s.client != nil {
	pt, err := influxdb.NewPoint(STATS_MEASUREMENT_NAME, tags, values, time.Now())
	if err != nil {
		log.Errorf("[TSDB] NewPoint failed. Error: %v", err)
		return err
	}

	s.rwl.Lock()
	defer s.rwl.Unlock()

	// see also https://github.com/influxdata/influxdb/blob/master/client/v2/example_test.go
	if s.bp == nil {
		// cmdr.SetDefault("server.tsdb.current.database", common.AppName)
		dbName := vxconf.GetStringR("server.tsdb.current.database", conf.AppName)
		// Create a new point batch
		s.bp, _ = influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
			Database:  dbName,
			Precision: "ns",
		})
	}

	s.bp.AddPoint(pt)
	// log.Debugf(">> [%d points]", len(s.bp.Points()))

	if len(s.bp.Points()) >= vxconf.GetIntR("server.tsdb.current.flush", 500) {
		go s.FlushStats()
	}
	// }
	return nil
}

func (s *influxClient) Query(ql string) (data interface{}, err error) {
	// cmdr.SetDefaultR("server.tsdb.current.database", common.AppName)
	dbName := vxconf.GetStringR("server.tsdb.current.database", conf.AppName)
	ql = strings.Replace(ql, ":m:", STATS_MEASUREMENT_NAME, -1)
	// log.Debugf("query db '%s'.'%s'.", dbName, STATS_MEASUREMENT_NAME)
	q := influxdb.NewQuery(ql, dbName, "ns")

	c := s.ensureTsdbClient()
	if c == nil {
		return
	}

	if response, e := c.Query(q); e == nil {
		if response.Error() != nil {
			return data, response.Error()
		}
		data = response.Results
	} else {
		err = e
	}
	return
}

const (
	// API 调用动作：call/invoke, break, open, close, down, up, ...
	ACTION_INVOKE = "call"
	ACTION_BREAK  = "break" // 意外中断了
	ACTION_OPEN   = "open"  // 重新连接上了，已经正确连接了
	ACTION_CLOSE  = "close" // 主动关闭了
	ACTION_DOWN   = "down"  // 下线了
	ACTION_UP     = "up"    // 上线了

	STATS_MEASUREMENT_NAME = "api_stats"

	DEFAULT_INFLUXDB_SERVER = "http://localhost:8086"

	MODE_HTTP = "HTTP"
	MODE_UDP  = "UDP"
)
