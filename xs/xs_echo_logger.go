/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"encoding/json"
	"github.com/labstack/echo"
	echolog "github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type CustomLogging struct {
	prefix string //
}

func stringToLevel(s string) echolog.Lvl {
	s = strings.ToUpper(s)
	switch s {
	case "DEBUG":
		return echolog.DEBUG
	case "INFO":
		return echolog.INFO
	case "WARN":
		return echolog.WARN
	case "ERROR":
		return echolog.ERROR
	default:
		return echolog.OFF
	}
}

func NewEchoLogger() echo.Logger {
	// fmt.Println("Using new echo logger / wrapped by CustomLogging.")
	// logex.InitLogger()
	return &CustomLogging{}
}

func (s *CustomLogging) Output() io.Writer {
	return os.Stderr
}
func (s *CustomLogging) SetOutput(w io.Writer) {
	log.SetOutput(w)
}
func (s *CustomLogging) Prefix() string {
	return s.prefix
}
func (s *CustomLogging) SetPrefix(p string) {
	s.prefix = p
}
func (s *CustomLogging) Level() echolog.Lvl {
	return echolog.INFO
}
func (s *CustomLogging) SetLevel(v echolog.Lvl) {
	switch v {
	case echolog.DEBUG:
		log.SetLevel(log.DebugLevel)
	case echolog.INFO:
		log.SetLevel(log.InfoLevel)
	case echolog.WARN:
		log.SetLevel(log.WarnLevel)
	case echolog.ERROR:
		log.SetLevel(log.ErrorLevel)
	case echolog.OFF:
		// WARN 一旦关闭了，没有实现重新打开的特征
		log.SetOutput(ioutil.Discard)
	}
}
func (s *CustomLogging) SetHeader(h string) {

}
func (s *CustomLogging) Print(i ...interface{}) {
	log.Info(i...)
}
func (s *CustomLogging) Printf(format string, args ...interface{}) {
	log.Infof(format, args...)
}
func (s *CustomLogging) Printj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Infof("%s", string(b))
	}
}
func (s *CustomLogging) Debug(i ...interface{}) {
	log.Debug(i...)
}
func (s *CustomLogging) Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
func (s *CustomLogging) Debugj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Debugf("%s", string(b))
	}
}
func (s *CustomLogging) Info(i ...interface{}) {
	log.Info(i...)
}
func (s *CustomLogging) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}
func (s *CustomLogging) Infoj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Infof("%s", string(b))
	}
}
func (s *CustomLogging) Warn(i ...interface{}) {
	log.Warning(i...)
}
func (s *CustomLogging) Warnf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}
func (s *CustomLogging) Warnj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Warnf("%s", string(b))
	}
}
func (s *CustomLogging) Error(i ...interface{}) {
	log.Error(i...)
}
func (s *CustomLogging) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
func (s *CustomLogging) Errorj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Errorf("%s", string(b))
	}
}
func (s *CustomLogging) Fatal(i ...interface{}) {
	log.Fatal(i...)
}
func (s *CustomLogging) Fatalj(j echolog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("error on marshalling json object: %v", j)
	} else {
		log.Fatalf("%s", string(b))
	}
}
func (s *CustomLogging) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
func (s *CustomLogging) Panic(i ...interface{}) {
	s.Fatal(i...)
	panic(i)
}
func (s *CustomLogging) Panicj(j echolog.JSON) {
	s.Fatal(j)
	panic(j)
}
func (s *CustomLogging) Panicf(format string, args ...interface{}) {
	s.Fatalf(format, args...)
	panic(nil)
}
