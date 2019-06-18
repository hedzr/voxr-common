/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cql

import (
	"github.com/gocql/gocql"

	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/labstack/gommon/log"
)

func (s *Cassandra) init(configItem *dbi.ConfigItem) dbi.DbBackend {
	s.C = gocql.NewCluster(configItem.Hosts...)
	s.C.Keyspace = configItem.Database
	s.C.Consistency = gocql.Quorum
	if configItem.DisableInitialHostLookup {
		// 本机开发和调试模式下，不要查找和刷新远程服务器的有效集群节点主机列表
		s.C.DisableInitialHostLookup = true
	}
	return s
}

func (s *Cassandra) Open() (session dbi.DbSession) {
	sss, err := s.C.CreateSession()
	if err != nil {
		log.Warnf("CreateSession() failed: %v", err)
		return nil
	}
	ss := &CassandraSession{sss}
	return ss
}

func (s *Cassandra) CloseAll() {
	return
}

func New(configItem *dbi.ConfigItem) (backend dbi.DbBackend) {
	c := &Cassandra{}
	c.init(configItem)
	return c
}
