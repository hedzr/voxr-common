/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"fmt"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/kvs/consul"
	"github.com/hedzr/voxr-common/kvs/etcd"
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"runtime"
)

func (s *Registrar) Open() {
	if s.Enabled == false {
		return
	}

	if s.IsOpen() {
		return
	}

	runMode := vxconf.GetStringR("runmode", "prod")
	s.Env = runMode

	switch s.Source {
	case "etcd":
		cfg := s.Etcd[s.Env]
		if len(cfg.Root) == 0 {
			cfg.Root = etcd.DEFAULT_ROOT_KEY
		}
		if addr := os.Getenv("REG_ADDR"); len(addr) > 0 {
			cfg.Peers = addr
		}
		logrus.Debugf("using cfg.addr=%s, REG_ADDR=%s", cfg.Peers, os.Getenv("REG_ADDR"))
		s.store = etcd.New(cfg)
	case "consul":
		cfg := s.Consul[s.Env]
		if addr := os.Getenv("REG_ADDR"); len(addr) > 0 {
			cfg.Addr = addr
		}
		logrus.Debugf("using cfg.addr=%s, REG_ADDR=%s", cfg.Addr, os.Getenv("REG_ADDR"))
		s.store = consul.New(cfg)
	}
	// return s.Store
}

func (s *Registrar) Close() {
	if s.store != nil {
		s.store.Close()
		s.store = nil
	}
}

func (s *Registrar) IsOpen() bool {
	return s.store != nil
}

func (s *Registrar) GetStore() store.KVStore {
	return s.store
}

// func toMap(m interface{}) (ret map[string]interface{}) {
// 	ret = make(map[string]interface{})
// 	if m != nil {
// 		if v, ok := m.(map[string]interface{}); ok {
// 			return v
// 		}
// 	}
// 	return
// }
//
// func toMapS(m interface{}) (ret map[string]string) {
// 	ret = make(map[string]string)
// 	if m != nil {
// 		if v, ok := m.(map[string]interface{}); ok {
// 			return toMapM(v)
// 		}
// 	}
// 	return
// }
//
// func toMapM(m map[string]interface{}) (ret map[string]string) {
// 	ret = make(map[string]string)
// 	for k, v := range m {
// 		switch reflect.ValueOf(v).Kind() {
// 		case reflect.Array, reflect.Slice, reflect.Map:
// 			ret[k] = fmt.Sprintf("%v", v)
// 		default:
// 			if v == nil {
// 				ret[k] = ""
// 			} else {
// 				ret[k] = fmt.Sprintf("%v", v)
// 			}
// 		}
// 	}
// 	return
// }

// SvrRecordRegister 注册服务到登记库中
func (s *Registrar) SvrRecordRegister(serviceName, version string) (err error) {
	if !s.Enabled || !s.IsOpen() {
		return fmt.Errorf("Disabled or Non-open registrar source: '%s'", s.Source)
	}

	switch s.Source {
	case "etcd", "consul":
		if r, ok := s.store.(store.Registrar); ok {
			ip, port, err := tool.LookupHostInfo()
			if err != nil {
				return err
			}

			tags := vxconf.GetStringSliceR("server.serviceTags", nil)
			meta := vxconf.GetMapR("server.serviceMeta", nil)
			id := vxconf.GetStringR("server.id", "")

			// or: try https://github.com/pbnjay/memory
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			var ds = tool.DiskUsage("/")
			tags = append(tags, fmt.Sprintf("version:%v", conf.Version))
			tags = append(tags, fmt.Sprintf("cpu:%v", runtime.NumCPU()))
			tags = append(tags, fmt.Sprintf("memory:%vM", mb(memory.TotalMemory())))
			tags = append(tags, fmt.Sprintf("disk:%vG", gb(ds.All)))
			tags = append(tags, fmt.Sprintf("os:%v", runtime.GOOS))
			tags = append(tags, fmt.Sprintf("arch:%v", runtime.GOARCH))
			meta["id"] = id
			meta["pid"] = fmt.Sprintf("%v", os.Getpid())
			meta["ppid"] = fmt.Sprintf("%v", os.Getppid())
			meta["cpu"] = fmt.Sprintf("%v", runtime.NumCPU())
			meta["memory"] = fmt.Sprintf("%vM", mb(memory.TotalMemory()))
			meta["memory-usage"] = fmt.Sprintf("alloc=%vM,total.alloc=%vM,sys=%vM,frees=%vM", mb(m.Alloc), mb(m.TotalAlloc), mb(m.Sys), mb(m.Frees))
			meta["os"] = fmt.Sprintf("%v", runtime.GOOS)
			meta["arch"] = fmt.Sprintf("%v", runtime.GOARCH)
			meta["disk"] = fmt.Sprintf("%vG", gb(ds.All))
			meta["disk-usage"] = fmt.Sprintf("free=%vG,used=%vG", gb(ds.Free), gb(ds.Used))

			logrus.Debugf("Register as service: %v, %v, %v, %v, %v", serviceName, id, version, ip, port)
			return r.Register(serviceName, id, version, ip, port, s.TTL, tags, meta, nil)
		} else {
			return fmt.Errorf("Unknown registrar (BAD developer and BAD CODE!)")
		}
	}
	return fmt.Errorf("Unknown registrar source: '%s'", s.Source)
}

func gb(size uint64) (ret uint64) {
	ret = size / 1024 / 1024 / 1024
	return
}

func mb(size uint64) (ret uint64) {
	ret = size / 1024 / 1024
	return
}

func (s *Registrar) SvrRecordDeregister(serviceName string, all bool) (err error) {
	if !s.Enabled || !s.IsOpen() {
		return nil
	}

	switch s.Source {
	case "etcd", "consul":
		if r, ok := s.store.(store.Registrar); ok {
			ip, port, err := tool.LookupHostInfo()
			if err != nil {
				return err
			}

			if all {
				return r.DeregisterAll(serviceName)
			} else {
				id := vxconf.GetStringR("server.id", "")
				return r.Deregister(serviceName, id, ip, port)
			}
		} else {
			return fmt.Errorf("Unknown registrar (BAD developer and BAD CODE!)")
		}
	}
	return fmt.Errorf("Unknown registrar source: '%s'", s.Source)
}

// SvrRecordResolver 返回单一一条可用服务记录（总是返回首条）
// return net.IPv4zero, 0    means there are no more targets to be found
func (s *Registrar) SvrRecordResolver(serviceName, ver, what string) (net.IP, uint16, string) {
	switch s.Source {
	case "etcd", "consul":
		if r, ok := s.store.(store.Registrar); ok {
			return r.NameResolver(serviceName, ver, what)
		} else {
			logrus.Fatalf("Unknown registrar (BAD developer and BAD CODE!)")
		}
	}
	return net.IPv4zero, 0, ""
}

// SvrRecordResolverAll 返回全部可用的服务记录
//  return all addresses if the service
// for ETCD, `what` can be "addr", "grpc"
// for CONSUL, `what` should be empty string now
func (s *Registrar) SvrRecordResolverAll(serviceName string, what string) []*store.ServiceRecord {
	switch s.Source {
	case "etcd", "consul":
		if r, ok := s.store.(store.Registrar); ok {
			return r.NameResolverAll(serviceName, what)
		} else {
			logrus.Fatalf("Unknown registrar (BAD developer and BAD CODE!)")
		}
	}
	return nil
}
