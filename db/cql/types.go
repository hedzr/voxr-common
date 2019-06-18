/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cql

import (
	"github.com/gocql/gocql"
	"github.com/hedzr/voxr-common/db/dbi"
)

type Cassandra struct {
	C *gocql.ClusterConfig
}

type CassandraSession struct {
	Instance *gocql.Session
}

type CassandraCRUD struct {
	TableName  string
	PrimaryKey []string
	Model      dbi.DbModel
	Session    *CassandraSession
}

type CassadraModel struct {
	// never used
}

const (
	CQL_KEY               = "cql"
	CQL_DEFAULT_VALUE_KEY = "default.cql"
	CQL_BEFORE_KEY        = "before"
	CQL_ENC_BCRYPT        = "bcrypt"
)
