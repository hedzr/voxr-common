/*
 * Copyright © 2019 Hedzr Yeh.
 */

package dbi

import "github.com/labstack/echo"

type (
	ConfigItem struct {
		Hosts                    []string `json:"hosts" yaml:"hosts"`
		Username                 string   `json:"username,omitempty" yaml:"username,omitempty"`
		Password                 string   `json:"password,omitempty" yaml:"password,omitempty"`
		Database                 string   `json:"database,omitempty" yaml:"database,omitempty"`
		ReplicaSet               string   `json:"replicaSet,omitempty" yaml:"replicaSet,omitempty"`
		DisableInitialHostLookup bool     `json:"disableInitialHostLookup,omitempty" yaml:"disableInitialHostLookup,omitempty"`
		Description              string   `json:"desc,omitempty" yaml:"desc,omitempty"`
		e                        *echo.Echo
	}

	// BackendItem struct {
	// 	Configs map[string]ConfigItem    `json:"configs" yaml:"configs"`
	// }

	WhatEnv map[string]*ConfigItem

	WhatDB struct {
		CurrentBackend string             `json:"backend" yaml:"backend"`
		CurrentEnv     string             `json:"env" yaml:"env"`
		Backends       map[string]WhatEnv `json:"backends" yaml:"backends"`
		Debug          bool               `json:"debug" yaml:"debug"`
	}

	Config struct {
		DB WhatDB `json:"db" yaml:"db"`
		// 后端的实例
		backend DbBackend `json:"-" yaml:"-"`
	}
)

var (
	DefaultDBConfig = Config{
		WhatDB{
			CurrentBackend: "couchdb",
			CurrentEnv:     "dev",
			Backends: map[string]WhatEnv{
				"couchdb": {
					// Items: map[string]ConfigItem{
					"dev": &ConfigItem{
						Hosts: []string{
							"localhost:5984",
						},
						Username: "admin",
						Password: "safe,2017",
						Database: "test",
					},
					// },
				},
			},
			Debug: true,
		},
		nil,
	}
)

type (
	DbModel interface {
	}

	DbCRUD interface {
		Insert() (rows int, err error)
		Update() (rows int, err error)
		// Get a single record by PK
		Get() (err error)
		Delete() bool
		Exists() bool

		// if condition checked then skip (or update the Model)
		// if unchecked (not exists), insert as a new record.
		InsertOrUpdate(conditions ...string) bool
		// InsertOrUpdate(conditions string, params ...interface{}) (rows int, err error)

		// Get a single record by conditions and params
		GetOne(conditions string, params ...interface{}) bool
		Query(results *[]interface{}, conditions string, params ...interface{}) *[]interface{}

		UpdateWith(conditionFields ...string) bool
		ExistsWith(conditionFields ...string) bool

		ExistsBy(conditions string, params ...interface{}) bool
		DeleteBy(conditions string, params ...interface{}) bool
	}

	DbSession interface {
		GetCRUD(model DbModel, tableName string, pkColumns ...string) (crud DbCRUD)
		Close()
	}

	DbBackend interface {
		// Init(configItem ConfigItem) (DbBackend)

		// must use the defer pattern:
		//   defer session.Close()
		Open() (session DbSession)
		// no use
		CloseAll()
	}
)

func (s *ConfigItem) SetEcho(e *echo.Echo) { s.e = e }
func (s *ConfigItem) GetEcho() *echo.Echo  { return s.e }

func (s *Config) GetBackend() DbBackend { return s.backend }
func (s *Config) SetBackend(b DbBackend) {
	if s.backend != nil {
		s.backend.CloseAll()
		s.backend = nil
	}
	s.backend = b
}

func (s *Config) IsValid() bool {
	return len(s.DB.CurrentBackend) > 0 && len(s.DB.CurrentEnv) > 0
}
