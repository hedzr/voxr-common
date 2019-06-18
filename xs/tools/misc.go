/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tools

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo"
	// "gopkg.in/mgo.v2"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	// "github.com/hedzr/voxr-common/the-server/log"
	log "github.com/sirupsen/logrus"
)

func DefaultPidPath(serviceName string) string {
	pidPath := "/var/run/${Service.Name}/${Service.Name}.pid" // c.String("pid")
	// fmt.Printf("pid path=%s\n", pidPath)
	if strings.Contains(pidPath, "$") {
		pidPath = strings.Replace(pidPath, "${Service.Name}", serviceName, -1)
	}
	return pidPath
}

func CreatePidFile(e *echo.Echo, serviceName string) error {
	pidPath := DefaultPidPath(serviceName)
	pidDir := path.Dir(pidPath)
	fmt.Printf("pid path=%s\n pid dir=%s\napp name=%s\n", pidPath, pidDir, serviceName)

	err := os.MkdirAll(pidDir, 0777)
	if err != nil {
		str := fmt.Sprintf("%v", err)
		log.Errorf("creating PID folder '%s': %s", pidDir, str)
		return err
	}
	// fmt.Printf("dir created: '%s'", pidDir)
	f, err := os.Create(pidPath)
	if err != nil {
		log.Errorf("creating PID file '%s': %v", pidPath, err)
		return err
	}
	// fmt.Printf("pid file created: '%s'", pidPath)
	pid := os.Getpid()
	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Errorf("writing PID file '%s': %v", pidPath, err)
		return err
	}

	log.Infof("'%s' running at :%d, as #%d, with '%s'.\n", serviceName, 0, pid, pidPath)
	return nil
}

func ReadContent(filename string) string {
	buf := bytes.NewBuffer(nil)
	// for _, filename := range filenames {
	f, err := os.Open(filename) // Error handling elided for brevity.
	if err != nil {
		return "0"
	}
	_, err = io.Copy(buf, f) // Error handling elided for brevity.
	if err != nil {
		return "0"
	}
	f.Close()
	// }
	return string(buf.Bytes())
}

func showRequestInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
	fmt.Fprintf(w, "Host: %s\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "RequestURI: %q\n", r.RequestURI)
	fmt.Fprintf(w, "URL: %#v\n", r.URL)
	fmt.Fprintf(w, "Body.ContentLength: %d (-1 means unknown)\n", r.ContentLength)
	fmt.Fprintf(w, "Close: %v (relevant for HTTP/1 only)\n", r.Close)
	fmt.Fprintf(w, "TLS: %#v\n", r.TLS)
	fmt.Fprintf(w, "\nHeaders:\n")
	r.Header.Write(w)
}

// func InitMongo(e *echo.Echo) *mgo.Session {
// 	// Database connection
// 	//var hosts []string //c.String("mongo.hosts")}
// 	//username, password, database, replica := "","","",""
// 	//var err error
//
// 	env := vxconf.GetStringSliceR("server.db.env")
// 	prefix := fmt.Sprintf("server.db.backends.%s.mongodb", env)
//
// 	hosts := vxconf.GetStringSliceR(prefix + "hosts")
// 	username := vxconf.GetStringR(prefix + "username")
// 	password := vxconf.GetStringR(prefix + "password")
// 	database := vxconf.GetStringR(prefix + "database")
// 	replica := vxconf.GetStringR(prefix + "replica")
//
// 	if len(hosts) > 0 {
// 		//session, err := mgo.Dial("mongodb://127.0.0.1:27017,localhost:27018,localhost:27019/?replicaSet=foo")
// 		// mongodb://localhost:27018/?replicaSet=rsdev&connect=replicaSet
// 		//session, err := mgo.Dial(hosts)
// 		session, err := mgo.DialWithInfo(&mgo.DialInfo{
// 			Addrs:          hosts,
// 			Username:       username,
// 			Password:       password,
// 			Database:       database,
// 			ReplicaSetName: replica,
// 		})
//
// 		if err != nil {
// 			e.Logger.Fatalf("Cannot connect to mongodb: %v", err) // will exit
// 		}
//
// 		//session.SetMode(mgo.Monotonic, true)
// 		//
// 		//// Drop Database
// 		//if IsDrop {
// 		//	err = session.DB("test").DropDatabase()
// 		//	if err != nil {
// 		//		panic(err)
// 		//	}
// 		//}
//
// 		// Create indices
// 		scopy := session.Copy()
// 		defer scopy.Close()
// 		if err = scopy.DB(database).C("users").EnsureIndex(mgo.Index{
// 			Key:    []string{"email"},
// 			Unique: true,
// 		}); err != nil {
// 			e.Logger.Fatalf("Cannot access mongodb DB '%s' or Collection 'users': %v", database, err) // will exit
// 		}
//
// 		var n int
// 		db := session.Clone()
// 		defer db.Close()
// 		if n, err = db.DB(database).C("users").Count(); err != nil {
// 			e.Logger.Fatalf("Cannot access mongodb DB '%s' or Collection: %v", database, err) // will exit
// 		}
//
// 		e.Logger.Info("MongoDB connected. n = %d (count of the users)\n", n)
// 		return session
// 	} else {
// 		e.Logger.Warn("MongoDB was DISABLED (since 'mongo.url' is empty).")
// 	}
// 	return nil
// }
