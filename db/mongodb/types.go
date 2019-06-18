/*
 * Copyright © 2019 Hedzr Yeh.
 */

package mongodb

import (
	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
)

type Mgo struct {
	// C *gocql.ClusterConfig
	e            *echo.Echo
	c            *dbi.ConfigItem
	firstSession *MgoSession
	sessions     []*MgoSession
}

type MgoSession struct {
	Instance *mgo.Session
	mgo      *Mgo
}

type MgoCRUD struct {
	TableName  string
	PrimaryKey []string
	Model      dbi.DbModel
	Session    *MgoSession
}

type MgoModel struct {
	// never used
}
