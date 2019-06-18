/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

import (
	"errors"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
)

// ExternalIP try to find the internet public ip address.
// it works properly in aliyun network.
//
// TODO only available to get IPv4 address.
//
// ExternalIP 尝试获得LAN地址.
// 对于aliyun来说，由于eth0是LAN地址，因此此函数能够正确工作；
// 对于本机多网卡的情况，通常这个函数的结果是正确的；
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// ThisHostname 返回本机名
func ThisHostname() string {
	name, err := os.Hostname()
	if err != nil {
		logrus.Warnf("WARN: %v", err)
		return "host"
	}
	return name
}

// ThisHost 返回当前服务器的LAN ip，通过本机名进行反向解析
func ThisHost() (ip net.IP) {
	ip = net.IPv4zero
	name, err := os.Hostname()
	if err != nil {
		logrus.Warnf("WARN: %v", err)
		return
	}
	logrus.Infof("detected os hostname: %s", name)
	ip, _, _ = hostInfo(name, 0)
	return
}

// LookupHostInfoOld 依据配置文件的 server.rpc_address 尝试解释正确的rpc地址，通常是IPv4的
func LookupHostInfoOld() (net.IP, int, error) {
	fAddr := vxconf.GetStringR("server.rpc_address", "0.0.0.0")
	fPort := vxconf.GetIntR("server.port", 2301)
	if fPort <= 0 || fPort > 65535 {
		fPort = DEFAULT_PORT
	}

	if len(fAddr) == 0 {
		name, err := os.Hostname()
		if err != nil {
			logrus.Warnf("WARN: %v", err)
			return net.IPv4zero, 0, err
		}
		logrus.Infof("detected os hostname: %s", name)
		fAddr = name
	}
	return hostInfo(fAddr, fPort)
}

func LookupHostInfo() (ip net.IP, port int, err error) {
	fAddr := vxconf.GetStringR("server.rpc_address", "0.0.0.0")
	port = vxconf.GetIntR("server.port", 2301)
	if port <= 0 || port > 65535 {
		port = DEFAULT_PORT
	}

	if len(fAddr) == 0 {
		addrs, _ := net.InterfaceAddrs()
		ips := make([]net.IP, len(prior)+1)
		var savedIP net.IP
		for _, addr := range addrs {
			// fmt.Println("IPv4/v6: ", addr)
			// addr.Network()
			if ipNet, ok := addr.(*net.IPNet); ok {
				if !ipNet.IP.IsUnspecified() {
					// 排除回环地址，广播地址
					// if ipNet.IP.IsGlobalUnicast() {
					// 	continue
					// }
					// // 排除公网地址
					if isLAN(ipNet.IP) {
						found := false
						for ix, x := range prior {
							if strings.HasPrefix(ipNet.IP.String(), x) {
								if ips[ix] == nil {
									ips[ix] = ipNet.IP
									found = true
									logrus.Debugf("    X %v", ipNet)
									break
								}
							}
						}
						if found == false {
							ips[len(prior)] = ipNet.IP
							// logrus.Debugf("    - %v | not found for prior list", ipNet)
						}
						// return ipNet.IP, port, nil
					} else {
						savedIP = ipNet.IP
						// logrus.Debugf("    . %v | savedIP", ipNet)
					}
				}
			}
		}

		for _, x := range ips {
			if !x.IsUnspecified() {
				ip = x
				return
			}
		}

		if savedIP != nil {
			logrus.Warnf("  A internet ipaddress found: '%s'; but we need LAN address; keep searching with findExternalIP().", savedIP.String())
		}

		// return findExternalIP(ipOrHost, port)

		return
	} else {
		return hostInfo(fAddr, port)
	}
}

var prior = []string{
	"192.168.0.", "192.168.1.", "10.", "172.16.", "172.17.", "172.29.",
}

func hostInfo(ipOrHost string, port int) (net.IP, int, error) {
	// macOS 可能会得到错误的主机名
	if strings.EqualFold("bogon", ipOrHost) {
		return findExternalIP(ipOrHost, port)
	}

	ip := net.ParseIP(ipOrHost)
	var savedIP net.IP
	if ip == nil || ip.IsUnspecified() {
		addrs, err := net.LookupHost(ipOrHost)
		if err != nil {
			logrus.Warnf("[kvs.store::hostInfo] Oops: LookupHost(): %v", err)
			return findExternalIP(ipOrHost, port)
		}

		ips := make([]net.IP, len(prior)+1)
		for _, addr := range addrs {
			ip2 := net.ParseIP(addr)
			if !ip2.IsUnspecified() {
				// 排除回环地址，广播地址
				// if ip2.IsGlobalUnicast() {
				// 	continue
				// }
				// // 排除公网地址
				if isLAN(ip2) {
					found := false
					for ix, x := range prior {
						if strings.HasPrefix(ip2.String(), x) {
							if ips[ix] == nil {
								ips[ix] = ip2
								found = true
								break
							}
						}
					}
					if found == false {
						ips[len(prior)] = ip2
					}
					// return ip2, port, nil
				} else {
					savedIP = ip2
				}
				logrus.Debugf("    x %v", ip2.String())
			} else {
				// allows ipV4/6 zero
				logrus.Debugf("    x %v", ip2.String())
				return findExternalIP(ip2.String(), port)
			}
		}
		for _, x := range ips {
			if x == nil {
				return hostInfo("localhost", port)
			}
			if !x.IsUnspecified() {
				return x, port, nil
			}
		}
		if savedIP != nil {
			logrus.Warnf("A internet ipaddress found: '%s'; but we need LAN address; keep searching with findExternalIP().", savedIP.String())
		}
		return findExternalIP(ipOrHost, port)
	}

	return ip, port, nil
	// return net.IPv4zero, 0, fmt.Errorf("cannot lookup 'server.rpc_address' or 'server.port'. cannot register myself.")
}

func isLAN(ip net.IP) bool {
	if ipv4 := ip.To4(); ipv4 != nil {
		if ipv4[0] == 192 && ipv4[1] == 168 {
			return true
		}
		if ipv4[0] == 172 && ipv4[1] == 16 {
			return true
		}
		if ipv4[0] == 10 {
			return true
		}
	} else {
		// TODO 识别IPv6的LAN地址段
	}
	return false
}

func findExternalIP(ipOrHost string, port int) (net.IP, int, error) {
	ip, err := ExternalIP()
	if err != nil {
		logrus.Errorf("Oops: findExternalIP(): %v", err)
		return net.IPv4zero, 0, err
	}
	logrus.Infof("use ip rather than hostname: %s", ip)
	// } else {
	// 	// NOTE 此分支尚未测试，由于macOS得到bogon时LookupHost() 必然失败，因此此分支应该是多余的
	// 	for _, a := range addrs {
	// 		fmt.Println(a)
	// 	}
	// }
	return net.ParseIP(ip), port, nil
}

const (
	DEFAULT_PORT = 6666
)
