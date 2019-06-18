/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type Mysql struct {
	// C *gocql.ClusterConfig
	e            *echo.Echo
	c            *dbi.ConfigItem
	dc           *dbi.Config
	firstSession *MysqlSession
	sessions     []*MysqlSession
}

type MysqlSession struct {
	Instance *gorm.DB
	mysql    *Mysql
}

type MysqlCRUD struct {
	TableName  string
	PrimaryKey []string
	Model      dbi.DbModel
	Session    *MysqlSession
}

type MysqlModel struct {
	// gorm.ModelStruct Model
}

func New(configItem *dbi.ConfigItem, dc *dbi.Config) (backend dbi.DbBackend) {
	c := &Mysql{}
	c.init(configItem, dc)
	return c
}

func (s *Mysql) init(configItem *dbi.ConfigItem, dc *dbi.Config) dbi.DbBackend {
	// s.C = gocql.NewCluster(configItem.Hosts...)
	// s.C.Keyspace = configItem.Database
	// s.C.Consistency = gocql.Quorum
	// if configItem.DisableInitialHostLookup {
	// 	// 本机开发和调试模式下，不要查找和刷新远程服务器的有效集群节点主机列表
	// 	s.C.DisableInitialHostLookup = true
	// }

	if len(configItem.Database) == 0 {
		configItem.Database = conf.AppName
	}

	s.c = configItem
	s.e = configItem.GetEcho()
	s.dc = dc
	s.sessions = []*MysqlSession{}

	if len(configItem.Hosts) > 0 {
		log.Infof("Connecting to mysql: %v", configItem.Hosts)
		// session, err := mgo.Dial("mongodb://127.0.0.1:27017,localhost:27018,localhost:27019/?replicaSet=foo")
		// mongodb://localhost:27018/?replicaSet=rsdev&connect=replicaSet
		// session, err := mgo.Dial(hosts)
		db := s.open()
		// defer db.Close()

		s.firstSession = &MysqlSession{db, s}
	}
	return s
}

func (s *Mysql) open() (db *gorm.DB) {
	cfg := mysql.NewConfig()
	cfg.Addr = s.c.Hosts[0]
	cfg.User = s.c.Username
	cfg.Passwd = s.c.Password
	cfg.DBName = s.c.Database

	cfg.ParseTime = true
	cfg.AllowNativePasswords = true
	dsn := cfg.FormatDSN()

	db, err := gorm.Open("mysql", dsn)
	// db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("Cannot connect to mysql: %v", err)
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	} else {
		log.Debug("mysql connected")
		if s.dc.DB.Debug {
			db.LogMode(true)
		}
	}
	return
}

func (s *Mysql) Open() (session dbi.DbSession) {
	// sss := s.firstSession.Instance.Copy() //s.C.CreateSession()
	db := s.open()
	ss := &MysqlSession{db, s}
	s.sessions = append(s.sessions, ss)
	// if err != nil {
	// 	s.e.Logger.Warnf("CreateSession() failed: %v", err)
	// 	return nil
	// }

	return ss
}

func (s *Mysql) CloseAll() {
	for _, s := range s.sessions {
		s.Close()
	}
	s.firstSession.Close()
	return
}

func (s *MysqlSession) DB() *gorm.DB {
	// return s.Instance.DB(s.mgo.c.Database)
	return s.Instance
}

func (s *MysqlSession) GetCRUD(model dbi.DbModel, tableName string, pkColumns ...string) (crud dbi.DbCRUD) {
	// TODO implement DbCRUD for mongodb
	return nil
}

func (s *MysqlSession) Close() {
	s.Instance.Close()
}
