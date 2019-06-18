/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mongodb

import (
	"github.com/hedzr/cmdr/conf"
	// "github.com/hedzr/voxr-common/log"
	"github.com/hedzr/voxr-common/db/dbi"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"time"
)

func New(configItem *dbi.ConfigItem) (backend dbi.DbBackend) {
	c := &Mgo{}
	c.init(configItem)
	return c
}

func (s *Mgo) init(configItem *dbi.ConfigItem) dbi.DbBackend {
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
	s.sessions = []*MgoSession{}

	if len(configItem.Hosts) > 0 {
		log.Infof("Connecting to mongodb: %v", configItem.Hosts)
		// session, err := mgo.Dial("mongodb://127.0.0.1:27017,localhost:27018,localhost:27019/?replicaSet=foo")
		// mongodb://localhost:27018/?replicaSet=rsdev&connect=replicaSet
		// session, err := mgo.Dial(hosts)
		session, err := mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:          configItem.Hosts,
			Username:       configItem.Username,
			Password:       configItem.Password,
			Database:       configItem.Database,
			ReplicaSetName: configItem.ReplicaSet,
			Timeout:        time.Second * 7,
		})

		if session == nil || err != nil {
			log.Fatalf("Cannot connect to mongodb: %v", err)
			// fmt.Printf("Cannot connect to mongodb: %v\n", err)
			panic(err) // will exit
		} else {
			log.Debug("mongodb connected")
			// fmt.Printf("mongodb connected: %v\n", session)
		}

		// session.SetMode(mgo.Monotonic, true)

		s.firstSession = &MgoSession{session, s}
	}
	return s
}

func (s *Mgo) Open() (session dbi.DbSession) {
	sss := s.firstSession.Instance.Copy() // s.C.CreateSession()
	ss := &MgoSession{sss, s}
	s.sessions = append(s.sessions, ss)
	// if err != nil {
	// 	s.e.Logger.Warnf("CreateSession() failed: %v", err)
	// 	return nil
	// }

	return ss
}

func (s *Mgo) CloseAll() {
	for _, s := range s.sessions {
		s.Close()
	}
	s.firstSession.Close()
	return
}

func (s *MgoSession) DB() *mgo.Database {
	return s.Instance.DB(s.mgo.c.Database)
}

func (s *MgoSession) GetCRUD(model dbi.DbModel, tableName string, pkColumns ...string) (crud dbi.DbCRUD) {
	// TODO implement DbCRUD for mongodb
	return nil
}

func (s *MgoSession) Close() {
	s.Instance.Close()
}
