/*
 * Copyright © 2019 Hedzr Yeh.
 */

package etcd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hedzr/voxr-common/kvs/id"
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
	// pb "go.etcd.io/etcd/etcdserver/etcdserverpb"
	// mvccpb "go.etcd.io/etcd/internal/mvcc/mvccpb"
)

type KVStoreEtcd struct {
	Client       *clientv3.Client
	e            *store.Etcdtool
	ctx          context.Context
	cancelFunc   context.CancelFunc
	lostConn     chan bool
	chGlobalExit chan bool
}

const (
	E3W_MODE         = false
	DIR_TAG          = "etcdv3_dir_$2H#%gRe3*t"
	DEFAULT_ROOT_KEY = "root"

	// etcd service registry root key (under DEFAULT_ROOT_KEY):
	SERVICE_REGISTRY_ROOT = "services"
)

var (
// defaultOpOption clientv3.OpOption
)

// ------------------------------------------------

func (s *KVStoreEtcd) SetDebug(enabled bool) {
	// debug = enabled
}

func (s *KVStoreEtcd) Open() {
	if s.Client != nil {
		if s.IsOpen() {
			return
		}

	} else {
		logrus.Warnf("etcd store has not been initialized.")
	}
}

func (s *KVStoreEtcd) Close() {
	if s.Client != nil {
		v("etcd client was been closing.")
		close(s.chGlobalExit)
		_ = s.Client.Close()
		s.Client = nil
	}
}

func (s *KVStoreEtcd) IsOpen() bool {
	if s.Client != nil {
		return true
	} else {
		return false
	}
}

func (s *KVStoreEtcd) GetRootPrefix() string {
	return s.e.Root
}

func (s *KVStoreEtcd) SetRoot(keyPrefix string) {
	s.e.Root = keyPrefix
}

func (s *KVStoreEtcd) Get(key string) string {
	// ctx, cancel := contextWithCommandTimeout(s.e.CommandTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}
	defer cancel() // s.cancelFunc()
	resp, err := s.Client.KV.Get(ctx, key)
	if err != nil {
		warnf("ERR: Get(%v) - %v", key, err)
		return ""
	}
	for _, ev := range resp.Kvs {
		return string(ev.Value)
	}
	return ""
}

func (s *KVStoreEtcd) GetYaml(key string) interface{} {
	// ctx, cancel := contextWithCommandTimeout(s.e.CommandTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}
	defer cancel() // s.cancelFunc()
	resp, err := s.Client.KV.Get(ctx, key)
	if err != nil {
		warnf("ERR: GetYaml(%v) - %v", key, err)
		return nil
	}

	// // build result
	var ret = make(map[string]interface{})
	for _, ev := range resp.Kvs {
		var v interface{}
		err := yaml.Unmarshal(ev.Value, &v)
		if err != nil {
			warnf("warn yaml unmarshal failed: %v", err)
		} else {
			if ev.Key == nil || strings.EqualFold(string(ev.Key), key) {
				return v
			}
			ret[string(ev.Key)] = v
		}
	}

	// and return the result
	return ret
}

func (s *KVStoreEtcd) GetPrefix(keyPrefix string) store.KvPairs {
	// ctx, cancel := contextWithCommandTimeout(s.e.CommandTimeout)
	if len(s.e.Root) > 0 {
		if len(keyPrefix) == 0 {
			keyPrefix = s.e.Root
		} else {
			keyPrefix = fmt.Sprintf("%s/%s", s.e.Root, keyPrefix)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	resp, err := s.Client.KV.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		warnf("ERR: GetPrefix(%v) - %v", keyPrefix, err)
		return nil
	}
	return &MyKvPairs{resp}
}

func (s *KVStoreEtcd) buildForDirectory(key string) {
	if E3W_MODE {
		sc := strings.Split(key, "/")
		sc = sc[0 : len(sc)-1]
		for ; len(sc) >= 1; sc = sc[0 : len(sc)-1] {
			dk := strings.Join(sc, "/")
			vf("  ** directory: %s\n", dk)
			if !strings.EqualFold(DIR_TAG, s.Get(dk)) {
				s.Put(dk, DIR_TAG)
			}
		}
	}
}

func (s *KVStoreEtcd) PutNX(key string, value string) error {
	key_old := key
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	_, err := s.Client.Put(ctx, key, value)
	if err != nil {
		return err
	}

	if !strings.EqualFold(DIR_TAG, value) {
		// e3w 需要这样的隐含规则，从而形成正确的目录结构。
		// etcd 本身对此过于随意。
		s.buildForDirectory(key_old)
	}
	return nil
}

func (s *KVStoreEtcd) Put(key string, value string) error {
	key_old := key
	if err := s.PutLite(key, value); err != nil {
		return err
	}
	if !strings.EqualFold(DIR_TAG, value) {
		// e3w 需要这样的隐含规则，从而形成正确的目录结构。
		// etcd 本身对此过于随意。
		s.buildForDirectory(key_old)
	}
	return nil
}

func (s *KVStoreEtcd) PutLite(key string, value string) error {
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	_, err := s.Client.Put(ctx, key, value)
	if err != nil {
		warnf("WARN: PutLite(%v, %v) - %v", key, value, err)
		return err
	}
	return nil
}

func (s *KVStoreEtcd) PutTTL(key string, value string, seconds int64) error {
	key_old := key
	if err := s.PutTTLLite(key, value, seconds); err != nil {
		return err
	}
	if !strings.EqualFold(DIR_TAG, value) {
		// e3w 需要这样的隐含规则，从而形成正确的目录结构。
		// etcd 本身对此过于随意。
		s.buildForDirectory(key_old)
	}
	return nil
}

func (s *KVStoreEtcd) PutTTLLite(key string, value string, seconds int64) (err error) {
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()

	resp, err2 := s.Client.Grant(ctx, seconds) // context.TO DO()
	if err2 != nil {
		err = err2
		if tool.IsCancelledOrDeadline(err) {
			return
		}
		// warnf("WARN: %v", err)
	}

	_, err = s.Client.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		if tool.IsCancelledOrDeadline(err) {
			return
		}
		// warnf("WARN: %v", err)
		return
	}
	return
}

func (s *KVStoreEtcd) LeaseKV(key string, value string, timeout int64) (err error) {
	const (
		ETCD_TRANSPORT_TIMEOUT = 7 * time.Second
		LEASE_TIME             = 10 // 10s
	)

	key_old := key
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ETCD_TRANSPORT_TIMEOUT)
	defer cancel()
	leaseResp, err := s.Client.Grant(ctx, timeout) // 租约时间设定为10秒
	if err != nil {
		return err
	}

	// ctx, cancel = context.WithTimeout(context.Background(), s.e.CommandTimeout)
	// defer cancel() //s.cancelFunc()
	var pr *clientv3.PutResponse
	pr, err = s.Client.Put(ctx, key, value, clientv3.WithLease(clientv3.LeaseID(leaseResp.ID)))
	if err != nil || pr == nil {
		if !tool.IsCancelledOrDeadline(err) {
			return
		}
	}
	logrus.Infof("put '%s' into '%s'", value, key)
	_, err = s.Client.KeepAlive(context.TODO(), leaseResp.ID)
	if !tool.IsCancelledOrDeadline(err) {
		return
	}

	if !strings.EqualFold(DIR_TAG, value) {
		// e3w 需要这样的隐含规则，从而形成正确的目录结构。
		// etcd 本身对此过于随意。
		s.buildForDirectory(key_old)
	}

	return
}

func (s *KVStoreEtcd) PutYaml(key string, value interface{}) (err error) {
	key_old := key
	if err = s.PutYamlLite(key, value); err != nil {
		return
	}

	// e3w 需要这样的隐含规则，从而形成正确的目录结构。
	// etcd 本身对此过于随意。
	s.buildForDirectory(key_old)
	return
}

func convert(input string) string {
	var buf bytes.Buffer
	for _, r := range input {
		if unicode.IsControl(r) {
			_, _ = fmt.Fprintf(&buf, "\\u%04X", r)
		} else {
			_, _ = fmt.Fprintf(&buf, "%c", r)
		}
	}
	return buf.String()
}

func (s *KVStoreEtcd) PutYamlLite(key string, value interface{}) (err error) {
	if s.Client == nil {
		err = errors.New("etcd not opened.")
		return
	}

	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}
	var b []byte
	b, err = yaml.Marshal(value)
	if err != nil {
		warnf("ERR: PutYamlLite(%v, %v) - %v", key, value, err)
		return
	}
	// infof("b: %v", string(b))
	// str := tool.UnescapeUnicode(b)
	str := string(b)

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	_, err = s.Client.Put(ctx, key, str)
	if err != nil {
		warnf("ERR: PutYamlLite(%v, %v) - %v", key, value, err)
		return
	}
	return
}

func (s *KVStoreEtcd) Delete(key string) bool {
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	resp, err := s.Client.Delete(ctx, key)
	if err != nil {
		warnf("ERR: Delete(%v) - %v", key, err)
		return false
	}
	infof("key '%s' was deleted: %v", key, resp)
	return true
}

func (s *KVStoreEtcd) DeletePrefix(keyPrefix string) bool {
	if len(s.e.Root) > 0 {
		if len(keyPrefix) == 0 {
			keyPrefix = s.e.Root
		} else {
			keyPrefix = fmt.Sprintf("%s/%s", s.e.Root, keyPrefix)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	defer cancel() // s.cancelFunc()
	resp, err := s.Client.Delete(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		warnf("ERR: DeletePrefix(%v) - %v", keyPrefix, err)
		return false
	}
	infof("key prefix '%s' was deleted: %v", keyPrefix, resp)
	return true
}

func (s *KVStoreEtcd) Exists(key string) bool {
	// etcd 没有key存在性检测的支持
	return false
}

// Watch 启动一个监视线程。etcd.Watch。
// stopCh被用于阻塞 blockFunc，并接收信号以结束该线程的阻塞。
// stopCh为nil时，阻塞线程失效，而且并不需要其存在，这一特殊场景仅仅适用于etcd的Watch机制。
func (s *KVStoreEtcd) Watch(key string, fn store.WatchFunc, stopCh chan bool) func() {
	if len(s.e.Root) > 0 {
		key = fmt.Sprintf("%s/%s", s.e.Root, key)
	}
	// ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	rch := s.Client.Watch(context.Background(), key)
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				// fmt.Printf("watch event - %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				// t := int(ev.Type)
				// et := store.Event_EventType(ev.Type)
				fn(store.Event_EventType(ev.Type), ev.Kv.Key, ev.Kv.Value)
			}
		}
		// close(rch)
	}()
	return func() {
		if stopCh != nil {
			<-stopCh
		}
	}
}

func (s *KVStoreEtcd) WatchPrefix(keyPrefix string, fn store.WatchFunc, stopCh chan bool) func() {
	if len(s.e.Root) > 0 {
		if len(keyPrefix) == 0 {
			keyPrefix = s.e.Root
		} else {
			keyPrefix = fmt.Sprintf("%s/%s", s.e.Root, keyPrefix)
		}
	}

	// 	ctx, cancel := context.WithTimeout(context.Background(), s.e.CommandTimeout)
	rch := s.Client.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	ch := stopCh
	if ch == nil {
		ch = s.chGlobalExit
	}
	go watchRunner(rch, fn, ch)
	// go func() {
	// 	//fmt.Printf("--> WatchPrefix(%s, ...) started.\n", keyPrefix)
	// 	for wresp := range rch {
	// 		//fmt.Printf("--> WatchPrefix(%s, ...) wakeup.\n", keyPrefix)
	// 		for _, ev := range wresp.Events {
	// 			//fmt.Printf("watch event - %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
	// 			//t := int(ev.Type)
	// 			//et := store.Event_EventType(ev.Type)
	// 			fn(store.Event_EventType(ev.Type), ev.Kv.Key, ev.Kv.Value)
	// 		}
	// 	}
	// 	//close(rch)
	// 	//fmt.Printf("--> WatchPrefix(%s, ...) stopped.\n\n", keyPrefix)
	// }()
	return func() {
		if stopCh != nil {
			logrus.Debugf("stopping etcd watcher routine")
			<-stopCh
			logrus.Debugf("stopped etcd watcher routine")
		}
		// close(rch)
	}
}

// Register in etcd, it register main restful service with ip+port, and register gRPC port implicitly.
func (s *KVStoreEtcd) Register(serviceName, id, version string, ipOrHost net.IP, port int, ttl int64, tags []string, meta map[string]interface{}, moreChecks api.AgentServiceChecks) error {
	// ip, port, err := store.LookupHostInfo()
	// if err != nil {
	// 	return err
	// }

	keyBase := fmt.Sprintf("%s/%s", SERVICE_REGISTRY_ROOT, serviceName)

	// addr := fmt.Sprintf("%s:%d", ipOrHost.String(), port)
	if len(id) == 0 {
		id = fmt.Sprintf("%s-%s-%d", serviceName, strings.Replace(ipOrHost.String(), ".", "-", -1), port)
	}

	// /root/services/<serviceName>:
	//    /peers/<serviceId>:
	//        addr: <ip:port>
	//        grpc: <port>
	//        ver: 1.x.1
	//        tags: ["",...]
	//        meta: {...}
	key := fmt.Sprintf("%s/peers/%s", keyBase, id)
	if ttl == 0 {
		ttl = 8
	}

	if err := s.Put(fmt.Sprintf("%s/ver", key), version); err != nil {
		logrus.Errorf("ERROR: register as service failed. %v", err)
		return err
	}

	if err := s.PutYaml(fmt.Sprintf("%s/tags", key), tags); err != nil {
		logrus.Errorf("ERROR: register as service failed. %v", err)
		return err
	}

	var keyRPC, grpcValue string
	var mainGrpc = fmt.Sprintf("server.grpc.%v", vxconf.GetStringR("server.grpc.main", "voxr-lite"))
	if len(mainGrpc) > 0 {
		var disabled bool
		grpcValue, id, disabled, _ = tool.LoadGRPCListen(mainGrpc)
		if !disabled {
			if len(mainGrpc) > 0 {
				meta["grpc"] = fmt.Sprintf("%v", grpcValue)
			}

			if err := s.PutYaml(fmt.Sprintf("%s/meta", key), meta); err != nil {
				logrus.Errorf("ERROR: register as service failed. %v", err)
				return err
			}

			keyRPC = fmt.Sprintf("%s/grpc", key)
		} else {
			logrus.Debugf("grpc service %v is disabled.", id)
		}
	} else {
		if err := s.PutYaml(fmt.Sprintf("%s/meta", key), meta); err != nil {
			logrus.Errorf("ERROR: register as service failed. %v", err)
			return err
		}
	}

	logrus.Debugf("    Register: key=%v, ver=%v, %v, %v, %v", key, s.Get(fmt.Sprintf("%s/ver", key)), ipOrHost, port, grpcValue)

	key = fmt.Sprintf("%s/addr", key)
	addrValue := fmt.Sprintf("%s:%d", ipOrHost.String(), port)

	go s.heartBeatRunner(key, keyRPC, addrValue, grpcValue, ttl)

	return nil
}

func (s *KVStoreEtcd) DeregisterAll(serviceName string) error {
	return s.deregister(serviceName, "", net.IPv4zero, 0)
}

func (s *KVStoreEtcd) Deregister(serviceName string, id string, ipOrHost net.IP, port int) error {
	return s.deregister(serviceName, id, ipOrHost, port)
}

func (s *KVStoreEtcd) deregister(serviceName string, id string, ipOrHost net.IP, port int) error {
	if s.IsOpen() {
		all := false

		// addr := fmt.Sprintf("%s:%d", ipOrHost.String(), port)
		if len(id) == 0 {
			if ipOrHost.IsUnspecified() || port == 0 {
				all = true
			} else {
				id = fmt.Sprintf("%s-%s-%d", serviceName, strings.Replace(ipOrHost.String(), ".", "-", -1), port)
			}
		} else {
			if ipOrHost.IsUnspecified() || port == 0 {
				all = true
			}
		}

		// 旧的算法比较ip和port以撤销服务注册；如果要求全部删除则无视ip和port
		// 新的算法已重新实现，比较id就好

		keyPrefix := fmt.Sprintf("%s/%s/peers", SERVICE_REGISTRY_ROOT, serviceName)
		if all {
			s.DeletePrefix(keyPrefix)
		} else {
			key := fmt.Sprintf("%s/%s", keyPrefix, id)
			s.DeletePrefix(key)
		}
		// peers := s.GetPrefix(keyPrefix)
		// for i := 0; i < peers.Count(); i++ {
		// 	ip := net.ParseIP(peers.Item(i).Key())
		// 	p, _ := strconv.Atoi(peers.Item(i).ValueString())
		//
		// 	del := all
		// 	if ! all && (ip.Equal(ipOrHost) && port == p) {
		// 		del = true
		// 	}
		//
		// 	if del {
		// 		key := fmt.Sprintf("%s/%s", keyPrefix, peers.Item(i).Key())
		// 		if ! s.Delete(key) {
		// 			log.Warnf("Delete etcd key '%s' failed", key)
		// 		}
		// 	}
		// }
	}
	return nil
}

// NameResolver etcd 服务发现框架：服务的每个实例注册自己到etcd, 且定时更新自己。
// 注册信息需要带有TTL，通常约定为 5s, 10s, 30s 等值。
func (s *KVStoreEtcd) NameResolver(serviceName, version, what string) (net.IP, uint16, string) {
	if s.IsOpen() {
		keyPrefix := fmt.Sprintf("%s/%s/peers", SERVICE_REGISTRY_ROOT, serviceName)
		peers := s.GetPrefix(keyPrefix)
		// 随机选择算法
		if peers.Count() > 0 {
			sel := rand.Intn(peers.Count())
			ip := net.ParseIP(peers.Item(sel).Key())
			port, err := strconv.Atoi(peers.Item(sel).ValueString())
			if err == nil {
				return ip, uint16(port), ""
			}
		}
	}
	return net.IPv4zero, 0, ""
}

// NameResolverAll 返回当前的全部可用服务记录列表
func (s *KVStoreEtcd) NameResolverAll(serviceName string, what string) (sr []*store.ServiceRecord) {
	if s.IsOpen() {
		keyPrefix := fmt.Sprintf("%s/%s/peers", SERVICE_REGISTRY_ROOT, serviceName)
		peers := s.GetPrefix(keyPrefix)
		if peers == nil {
			return
		}

		const addr = "/addr"
		const vers = "/ver"
		// const grpc = "/grpc"
		suffix := fmt.Sprintf("/%s", what)
		type A struct {
			ipBase, ip net.IP
			port       int16
			version    string
			whatValue  string
		}
		var b = make(map[string]*A)
		var ip net.IP
		var l int

		for sel := 0; sel < peers.Count(); sel++ {
			item := peers.Item(sel)
			k := item.Key()

			if l == 0 {
				l = strings.Index(k, keyPrefix) + len(keyPrefix)
			}
			ks := k[l:]
			if len(ks) == 0 {
				continue
			}
			ks = strings.Split(ks[1:], "/")[0]

			if _, ok := b[ks]; !ok {
				b[ks] = &A{nil, nil, 0, "", ""}
			}
			// log.Debugf("    - checking for, key:%v | %v, value:%v", k, ks, item.ValueString())
			if strings.HasSuffix(k, addr) {
				parts := strings.Split(item.ValueString(), ":")
				b[ks].ipBase = net.ParseIP(parts[0])
				// log.Debugf("      ip hit: %v", b[ks].ipBase)
			} else if strings.HasSuffix(k, vers) {
				b[ks].version = item.ValueString()
				// log.Debugf("      version hit: %v", b[ks].version)
			} else if strings.HasSuffix(k, suffix) {
				b[ks].whatValue = item.ValueString()
			}
		}

		for _, v := range b {
			if len(v.whatValue) == 0 {
				// log.Debugf("        %v has empty %v setting.", k, what)
				continue
			}
			parts := strings.Split(v.whatValue, ":")
			s := parts[0]
			if len(s) == 0 || s == "0.0.0.0" || s == "::1" || s == "*" {
				ip = v.ipBase
			} else {
				ip = net.ParseIP(s)
			}
			port, err := strconv.Atoi(parts[1])
			if err == nil {
				// TODO build id for ipv6/ipv4
				id1 := id.GenerateServerRecordId(ip, port) // fmt.Sprintf("%s;%d", ip.String(), port)
				r := &store.ServiceRecord{IP: ip, Port: uint16(port), ID: id1, What: what, Version: v.version}
				sr = append(sr, r)
				// log.Debugf("        add service record：%v", *r)
			}
		}

	}
	return
}
