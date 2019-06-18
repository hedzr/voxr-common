/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultForwardersConfig = Config{
		Forwarders: []*FwdItem{},
		Skipper:    middleware.DefaultSkipper,
		// Level:   -1,
	}

	HeaderXServer = false

	ThisConfig Config
)

func GetAppConfig() *Config {
	return &ThisConfig
}

func Shutdown() {
	ThisConfig.Shutdown()
}

func (s *Config) Shutdown() {
	if s.Forwarders != nil {
		for _, fwdr := range s.Forwarders {
			if fwdr != nil {
				fwdr.Shutdown()
			}
		}
	}

	s.Registrar.Close()
}

func (s *Config) Init() {
	HeaderXServer = vxconf.GetBoolR("server.headerXServer", true)

	// fObj := vxconf.GetR("server.forwarders")
	// //fmt.Printf("Forwarders String Got: %v\n\n", fObj)
	//
	// b, err := yaml.Marshal(fObj)
	// check_panic(err)
	// s2 := string(b)
	// if gwk.Debug {
	// 	fmt.Printf("Forwarders Yaml Built: \nforwarders:\n%v\n\n", s2)
	// }
	// gwk.ThisConfig = gwk.DefaultForwardersConfig
	// gwk.ThisConfig.Forwarders = []*gwk.FwdItem{}
	// err = yaml.Unmarshal([]byte(s2), &gwk.ThisConfig.Forwarders)
	// check_panic(err)
	// //fmt.Printf("Forwarders Got #1: %v\n\n", thisConfig.Forwarders)

	// load the registrar configurations
	fObj := vxconf.GetMapR("server.registrar", nil)
	// fmt.Printf("registrar String Got: %v\n\n", fObj)

	b, err := yaml.Marshal(fObj)
	if err == nil {
		s2 := string(b)
		// logrus.Debugf("registrar Yaml Built: \n%v\n\n", s2)
		s.Registrar = Registrar{}
		err = yaml.Unmarshal([]byte(s2), &s.Registrar)
	}

	if err == nil {
		// Open the connection to registrar, enable `store`
		s.Registrar.Open()
	}
}
