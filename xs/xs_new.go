/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"github.com/hedzr/voxr-common/vxconf"
)

func New(serverImpl vxconf.CoolServer) vxconf.XsServer {
	return &echoServerImpl{
		e:    nil,
		cool: serverImpl,
	}
}
