/*
 * Copyright © 2019 Hedzr Yeh.
 */

package db

import (
	"fmt"
	"github.com/hedzr/voxr-common/db/cql"
	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/hedzr/voxr-common/db/mongodb"
	"github.com/hedzr/voxr-common/db/mysql"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/labstack/echo"
	"os"
	"strings"

	// 	y "github.com/ghodss/yaml"
	"github.com/couchbase/go-couchbase"
	log "github.com/sirupsen/logrus"

	"time"
)

var (
	thisConfig    *dbi.Config
	theConfigItem *dbi.ConfigItem
)

func check_panic(err error) {
	if err != nil {
		panic(err)
	}
}

func GetDbConfig() *dbi.Config {
	if thisConfig != nil && thisConfig.IsValid() {
		return thisConfig
	}
	GetDbConfigItem()
	return thisConfig
}

func GetDbConfigItem() *dbi.ConfigItem {
	if theConfigItem != nil && theConfigItem.Hosts != nil && len(theConfigItem.Hosts) > 0 {
		return theConfigItem
	}

	thisConfig = &dbi.Config{}
	if err := vxconf.LoadSectionTo("server.pub.deps.db", thisConfig); err != nil {
		log.Fatalf("error: %v", err)
	}

	runMode := vxconf.GetStringR("runmode", "devel")
	if len(runMode) > 0 {
		thisConfig.DB.CurrentEnv = runMode
	}

	// fObj := vxconf.GetR("server.db")
	// b, err := yaml.Marshal(fObj)
	// check_panic(err)
	// thisConfig = &dbi.Config{}
	// err = yaml.Unmarshal(b, &thisConfig.DB)
	// check_panic(err)
	// //log.Debugf("DB config Got #1: %v\n\n", thisConfig.DB)

	theConfigItem := thisConfig.DB.Backends[thisConfig.DB.CurrentBackend][thisConfig.DB.CurrentEnv]

	if s := os.Getenv("DB_ADDR"); len(s) > 0 {
		theConfigItem.Hosts = strings.Split(s, ",")
	}

	log.Infof("DB current config item Got: %v, '%v'", theConfigItem.Hosts, theConfigItem.Description)

	// env := vxconf.GetStringSliceR("server.db.env")
	// backendStr := vxconf.GetStringSliceR("server.db.backend")
	// prefix := fmt.Sprintf("server.db.backends.%s.%s", env, backendStr)
	// fObj = vxconf.GetR(prefix)
	//
	// b, err = yaml.Marshal(fObj)
	// check_panic(err)
	// //if true {
	// //	s2 := string(b)
	// //	fmt.Print("---")
	// //	fmt.Printf("Forwarders Yaml Built: \n%v\n\n", s2)
	// //	fmt.Print("--- END")
	// //}
	//
	// theConfigItem = dbi.ConfigItem{}
	// err = yaml.Unmarshal(b, &theConfigItem)
	// check_panic(err)
	// fmt.Printf("DB current config item Got #2: %v\n\n", theConfigItem)
	return theConfigItem
}

func GetBackend() dbi.DbBackend {
	return thisConfig.GetBackend()
}

func GetCurrentBackendString() string {
	return thisConfig.DB.CurrentBackend
}

func GetCurrentEnvString() string {
	return thisConfig.DB.CurrentEnv
}

func Init(e *echo.Echo) *dbi.Config {
	// if true {
	// 	thisConfig = config.DefaultDBConfig
	// 	bt, err := yaml.Marshal(thisConfig)
	// 	check_panic(err)
	// 	fmt.Print(string(bt))
	// 	fmt.Print("--- END")
	// }

	item := GetDbConfigItem()
	item.SetEcho(e)

	switch GetCurrentBackendString() {
	case "couchdb":
		thisConfig.SetBackend(initCouchDB(e, item))
	case "mongodb":
		thisConfig.SetBackend(initMongoDB(item))
	case "cassandra":
		thisConfig.SetBackend(initCassandraDB(e, item))
	case "mysql":
		thisConfig.SetBackend(initMysqlDB(e, item, thisConfig))
	}
	return thisConfig
}

func Close() {
	// item := GetDbConfigItem()
	// if thisConfig.GetBackend() != nil {
	// thisConfig.GetBackend().CloseAll()
	thisConfig.SetBackend(nil) // setBackend将会清除原有的backend的全部连接(CloseAll())
	// }
}

func initMysqlDB(e *echo.Echo, configItem *dbi.ConfigItem, dc *dbi.Config) dbi.DbBackend {
	if len(configItem.Hosts) > 0 {
		backend := mysql.New(configItem, dc)
		return backend
	} else {
		log.Warn("MySQL DB was DISABLED (since 'db...mysql.hosts' is empty '[]').")
	}
	return nil
}

func initCouchDB(e *echo.Echo, configItem *dbi.ConfigItem) dbi.DbBackend {
	for _, host := range configItem.Hosts {
		url := fmt.Sprintf("http://%s/", host)
		c, err := couchbase.Connect(url)
		if err != nil {
			log.Fatalf("Error connecting:  %v", err)
		}

		pool, err := c.GetPool(configItem.Database)
		if err != nil {
			log.Fatalf("Error getting pool:  %v", err)
		}

		bucket, err := pool.GetBucket("default")
		if err != nil {
			log.Fatalf("Error getting bucket:  %v", err)
		}

		bucket.Set("someKey", 0, []string{"an", "example", "list"})
	}
	return nil
}

func initMongoDB(configItem *dbi.ConfigItem) dbi.DbBackend {
	// tools.InitMongo(e)

	if len(configItem.Hosts) > 0 {
		backend := mongodb.New(configItem)
		return backend
	} else {
		log.Warn("MongoDB was DISABLED (since 'db...mongodb.hosts' is empty).")
	}
	return nil
}

type MxModel interface {
	Init()
}

type Users2 struct {
	// ID          gocql.UUID    `json:"id" cql:"id" bson:"_id,omitempty" default.cql:"uuid()"`
	Name        string    `json:"username" cql:"username" bson:"username"` // unique
	Email       string    `json:"email" cql:"email" bson:"email"`          // unique
	Mobile      string    `json:"mobile" cql:"mobile" bson:"mobile"`       // unique
	Password    string    `json:"password,omitempty" cql:"password" bson:"password" before:"bcrypt"`
	CreatedTime time.Time `json:"time_created" cql:"time_created" bson:"time_created" default.cql:"toTimestamp(now())"`
	UpdatedTime time.Time `json:"time_updated" cql:"time_updated" bson:"time_updated" default.cql:"toTimestamp(now())"`
	Token       string    `json:"token1,omitempty" cql:"token1" bson:"-"`
	Followers   []string  `json:"followers,omitempty" cql:"followers" bson:"followers,omitempty"`

	Blocked   bool `json:"blocked" cql:"s_blocked" bson:"email"`     // 临时禁用，管理员操作时
	Forbidden bool `json:"forbidden" cql:"s_forbidden" bson:"email"` // 自动规则所永久禁用
	Deleted   bool `json:"deleted" cql:"s_deleted" bson:"email"`     // 已经销户
}

func (s Users2) Init() {
	// s.ID = uuid
	// s.CreateTime = xxx
}

func initCassandraDB(e *echo.Echo, configItem *dbi.ConfigItem) dbi.DbBackend {
	backend := cql.New(configItem)
	// fmt.Printf("backend is : %v", backend)

	// uuid1 := uuid.NewV4()
	// uuid, err := gocql.UUIDFromBytes(uuid1.Bytes())
	// if err != nil {
	// 	panic(err)
	// }
	u := Users2{
		// ID:        uuid,
		Name:      "admin2",
		Password:  "dush9,ever!01x",
		Blocked:   false,
		Forbidden: false,
		Deleted:   false,
	}
	// fmt.Printf("using uuid %s\n", uuid1.String())

	session := backend.Open()
	if session == nil {
		panic("cannot connect with cassandra servers.")
	}
	defer session.Close()

	crud := session.GetCRUD(&u, "users", "username")

	crud.InsertOrUpdate("username=?", u.Name)
	// if ! crud.Exists() {
	// 	crud.Insert()
	// }

	if !crud.GetOne("username=?", u.Name) {
		panic("cannot GetOne().")
	}
	// fmt.Printf("Got Users2: id=%s\n", u.ID)

	u.Email = "admin2@example.com"
	crud.Update()

	u.Name = ""
	u.Email = ""
	u.Mobile = ""
	crud.Get()
	fmt.Printf("Got Users2: username=%s\n", u.Name)

	u2 := Users2{}
	crud2 := session.GetCRUD(&u2, "users", "username")
	if crud2.GetOne("username=?", "admin1") {
		crud2.Delete()
	} else {
		fmt.Printf("user '%s' NOT FOUND. Nothing to do.", "admin1")
	}

	return backend
}

func GetSession() {
	//
}
