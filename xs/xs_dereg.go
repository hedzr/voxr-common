/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"fmt"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/vxconf/gwk"
	"github.com/sirupsen/logrus"
)

func Deregister() (err error) {
	// forwarder.InitConfig()

	serviceName := vxconf.GetStringR("server.serviceName", conf.AppName)
	if len(serviceName) == 0 {
		logrus.Fatalf("serviceName must be specified!")
		return fmt.Errorf("serviceName must be specified")
	}

	all := vxconf.GetBoolR("server.registrar.deregister.all", true)
	logrus.Debugf("    deregister.all = %v", all)

	r := gwk.ThisConfig.Registrar
	r.Open()

	if err := r.SvrRecordDeregister(serviceName, all); err != nil {
		logrus.Fatalf("cannot deregister service from '%s': %v", gwk.ThisConfig.Registrar.Source, err)
		return err
	} else {
		logrus.Info("deregister ok.")
		return nil
	}
	return
}
