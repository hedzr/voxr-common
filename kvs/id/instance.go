/*
 * Copyright © 2019 Hedzr Yeh.
 */

package id

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
)

func GetInstanceId() string {
	return vxconf.GetStringR("server.id", "")
}

func GenerateInstanceId() {
	GenerateInstanceIdWithAppName(conf.AppName, &conf.ServerID)
}

func GenerateInstanceIdWithAppName(appName string, serverId *string) {
	*serverId = GetInstanceId()
	if len(*serverId) == 0 {
		rpc_addr, rpc_port, err := tool.LookupHostInfo()
		if err != nil {
			logrus.Fatalf("buildInstanceID: store.LookupHostInfo() failed. ERR: %v", err)
		}

		// id := fmt.Sprintf("%s-%s-%s-%d", appName, store.ThisHostname(), rpc_addr, rpc_port)
		id := fmt.Sprintf("%s:%s:%d", appName, rpc_addr, rpc_port)
		cmdr.Set("server.id", id)
		*serverId = id
		logrus.Debugf("buildInstanceID: id-generated = %v", *serverId)
	} else {
		logrus.Debugf("buildInstanceID: using id = %v", *serverId)
	}
}
