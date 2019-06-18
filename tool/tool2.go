/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/voxr-common/vxconf"
	"strconv"
	"strings"
)

var port int
var host string

func Port() int {
	return port
}

func IncPort() int {
	port++
	return port
}

func IncPortAndAddr(prefix string) (addr string) {
	addr = fmt.Sprintf("%s:%d", host, IncPort())
	cmdr.Set(fmt.Sprintf("%s.port", prefix), Port())
	return
}

// LoadHostDefinition loads the dependence info from config entry `server.deps.xxx`
// LoadHostDefinition("server.deps.apply")
func LoadHostDefinition(prefix string) (addr string) {
	addr = vxconf.GetStringR(fmt.Sprintf("%s.addr", prefix), "")
	host = vxconf.GetStringR(fmt.Sprintf("%s.host", prefix), "0.0.0.0")
	port = vxconf.GetIntR(fmt.Sprintf("%s.port", prefix), 2300)
	if len(host) > 0 && len(addr) == 0 {
		addr = fmt.Sprintf("%s:%d", host, port)
		// } else {
		// logrus.Debugf("prefix = %v, addr = %v", prefix, addr)
		// parts := strings.Split(addr, ":")
		// host = parts[0]
		// port, _ = strconv.Atoi(parts[1])
	}
	return
}

// LoadGRPCListen("server.grpc.apply")
func LoadGRPCListen(prefix string) (listen string, id string, disabled bool, port int) {
	listen = vxconf.GetStringR(fmt.Sprintf("%v.listen", prefix), "")
	id = vxconf.GetStringR(fmt.Sprintf("%v.id", prefix), "")
	disabled = vxconf.GetBoolR(fmt.Sprintf("%v.disabled", prefix), false)
	parts := strings.Split(listen, ":")
	port, _ = strconv.Atoi(parts[1])
	return
}

func IncGrpcListen(prefix string) (listen string, port int) {
	listen = vxconf.GetStringR(fmt.Sprintf("%v.listen", prefix), "")
	parts := strings.Split(listen, ":")
	p, _ := strconv.Atoi(parts[1])
	p++
	port = p
	listen = fmt.Sprintf("%s:%d", parts[0], p)
	cmdr.Set(fmt.Sprintf("%s.listen", prefix), listen)
	return
}

//
// what: "addr", "grpc", "health"
//
func FindService(what string) {
	// /
}
