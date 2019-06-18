/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/vxconf/gwk"
	"github.com/labstack/echo"
)

type RegistrarCallback func(r *gwk.Registrar, serviceName string)

var okRegistrarCallback RegistrarCallback
var changesHandler store.WatchFunc

func SetRegistrarOkCallback(okCallback RegistrarCallback) {
	okRegistrarCallback = okCallback
}
func SetRegistrarChangesHandler(handler store.WatchFunc) {
	changesHandler = handler
}

// registerAsService 注册自己到 registrar, 成为一个well-known的公共服务
func registerAsService(e *echo.Echo) {
	r := gwk.ThisConfig.Registrar
	// st := r.Store
	if !r.IsOpen() {
		if !r.Enabled {
			gwk.ThisConfig.Init()
			r = gwk.ThisConfig.Registrar
		}
		r.Open()
	}

	fServiceName := vxconf.GetStringR("server.serviceName", conf.AppName)
	fVersion := vxconf.GetStringR("server.version-sim", vxconf.GetStringR("version-sim", conf.Version))

	if err := r.SvrRecordRegister(fServiceName, fVersion); err != nil {
		e.Logger.Fatalf("cannot register as a '%s' service: %v", gwk.ThisConfig.Registrar.Source, err)
	} else {
		listenChanges(&r, fServiceName)
		if okRegistrarCallback != nil {
			okRegistrarCallback(&r, fServiceName)
		}
	}
}

var chStopListen chan bool

func listenChanges(r *gwk.Registrar, serviceName string) {
	if r.IsOpen() {
		chStopListen = make(chan bool)
		// fmt.Sprintf("services/%s", serviceName)
		r.GetStore().WatchPrefix("services", func(evType store.Event_EventType, key []byte, value []byte) {
			if changesHandler != nil {
				changesHandler(evType, key, value)
			}
		}, chStopListen)
	}
}
