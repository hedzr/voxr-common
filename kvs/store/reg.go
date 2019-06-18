/*
 * Copyright © 2019 Hedzr Yeh.
 */

package store

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/tool"
	"net"
	"strconv"
	"strings"
)

type (
	Registrar interface {
		// Register id省缺时，通过ip+port自动生成一个; 否则以id为准;
		// `ipOrHost` is a host:port string, `port` was ignored;
		// when `ipOrHost` is a IP string, `port` must be valid port number;
		//
		// for ETCD, ttl (in seconds) should be a valid time number;
		// for ETCD, it register gRPC port implicitly too.
		Register(serviceName, id, version string, ipOrHost net.IP, port int, ttl int64, tags []string, meta map[string]interface{}, moreChecks api.AgentServiceChecks) error
		// Deregister id省缺时，通过ip+port自动生成一个; 否则以id为准;
		// `ipOrHost` is a host:port string, `port` was ignored;
		// when `ipOrHost` is a IP string, `port` must be valid port number;
		Deregister(serviceName string, id string, ipOrHost net.IP, port int) error
		DeregisterAll(serviceName string) error
		// NameResolver return the addr field of the service
		// NameResolver(serviceName string) (net.IP, uint16)
		NameResolver(serviceName, version, what string) (ip net.IP, port uint16, versionHit string)
		// NameResolverAll return all addresses if the service
		// for ETCD, `what` can be "addr", "grpc"
		// for CONSUL, `what` should be empty string now
		NameResolverAll(serviceName string, what string) []*ServiceRecord
	}

	ServiceRecord struct {
		IP      net.IP
		Port    uint16 // RESTful port, or gRPC port,
		ID      string
		Version string
		What    string // 备用。多数时候为如下值：ADDR(RESTful), GRPC
		IsLocal bool   // 当此记录是一条预定义、静态记录时；区别于从服务注册中心取得的记录
	}
)

func (sr *ServiceRecord) String() (s string) {
	return fmt.Sprintf("%v:%v (%s) v%v", sr.IP, sr.Port, sr.ID, sr.Version)
}

func NewServiceRecord(addr string) *ServiceRecord {
	return NewServiceRecordWithVersion(addr, conf.Version)
}

// ":7001" => *ServiceRecord 建立静态记录
func NewServiceRecordWithVersion(addr, version string) *ServiceRecord {
	if strings.Contains(addr, ":") {
		parts := strings.Split(addr, ":")
		if parts[0] == "::1" || parts[0] == "0.0.0.0" || len(parts[0]) == 0 {
			if ip, err := tool.ExternalIP(); err == nil {
				if port, err := strconv.Atoi(parts[1]); err == nil {
					return &ServiceRecord{net.ParseIP(ip), uint16(port),
						fmt.Sprintf("%s;%d", ip, port),
						// id.GenerateServerRecordId(net.ParseIP(ip), port),
						version, "", true}
				}
			}
		} else {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				return &ServiceRecord{net.ParseIP(parts[0]), uint16(port),
					fmt.Sprintf("%s;%d", parts[0], port),
					// id.GenerateServerRecordId(net.ParseIP(parts[0]), port),
					version, "", true}
			}
		}
	}
	return nil
}

func (sr *ServiceRecord) IsLocalDefined() bool {
	return sr.IsLocal // sr.IP == nil && len(sr.What) > 0 && strings.Contains(sr.What, ":")
}

func (sr *ServiceRecord) Equal(other *ServiceRecord) (ok bool) {
	ok = sr.IP.Equal(other.IP) && sr.Port == other.Port // && sr.ID == other.ID && sr.What == other.What
	return
}
