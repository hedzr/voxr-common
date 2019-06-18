/*
 * Copyright © 2019 Hedzr Yeh.
 */

package etcd

import (
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *KVStoreEtcd) heartBeatAction(key, keyRPC, addrValue, grpcValue string, ttl int64) (err error) {
	// addrValue := fmt.Sprintf("%s:%d", ipOrHost.String(), port)

	// 在TTL失效阈值之内，完成心跳更新操作。etcd采用TTL机制，无需清除无效的服务注册表项。
	if err = s.PutTTL(key, addrValue, ttl+1); err != nil {
		logrus.Warnf("[heartBeatAction] BAD: service record put failed. %v", err)
		return
	}
	if len(keyRPC) > 0 {
		if err = s.PutTTL(keyRPC, grpcValue, ttl+1); err != nil {
			// logrus.Errorf("ERROR: register as service failed. %v", err)
			logrus.Warnf("[heartBeatAction] BAD: service record put failed. %v", err)
			return
		}
	}
	return
}

func (s *KVStoreEtcd) heartBeatRunner(key, keyRPC, addrValue, grpcValue string, ttl int64) {
	if err := s.heartBeatAction(key, keyRPC, addrValue, grpcValue, ttl); err != nil {
		if !tool.IsCancelled(err) {
			logrus.Errorf("ERROR: register as service failed. %v", err)
			return
		}
	}

	ticker := time.NewTicker(time.Duration(ttl+1) * time.Second)
	defer func() {
		ticker.Stop()
		// hub.unregister <- c
		// if err := c.conn.Close(); err != nil {
		// 	logrus.Warnf("error occurs at ws closing: %v", err)
		// }
	}()

	for {
		select {
		case tm := <-ticker.C:
			if err := s.heartBeatAction(key, keyRPC, addrValue, grpcValue, ttl); err != nil {
				// logrus.Errorf("ERROR: register as service failed. %v", err)
				// return
				if !tool.IsCancelledOrDeadline(err) {
					logrus.Warningf("ERROR: etcd service heartbeat (RESTful) failed at %v. %v", tm, err)
				} else {
					logrus.Infof("NOTICE: etcd heartbeat loop timeout/deadline at %v", tm)
					if !s.IsOpen() {
						s.Open()
					}
				}
			} else {
				var foreground = vxconf.GetBoolR("server.foreground", false)
				if noHeartbeatLog == false && foreground {
					// 心跳信息不要记录到日志中，因此这里不使用 logger
					logrus.Debugf("HEARTBEAT: etcd service heartbeat ok. per %d seconds.", ttl+1)
				}
			}

		case <-s.chGlobalExit:
			logrus.Warningf("etcd service heartbeat loop is existing...")
			return
		}
	}

	// for s.IsOpen() {
	// 	if err := s.heartBeatAction(key, keyRPC, addrValue, grpcValue, ttl); err != nil {
	// 		//logrus.Errorf("ERROR: register as service failed. %v", err)
	// 		//return
	// 		logrus.Warningf("ERROR: etcd service heartbeat (RESTful) failed. %v", err)
	// 		t := (ttl + 2) / 3
	// 		if t == 0 {
	// 			t = 1
	// 		}
	// 		time.Sleep(time.Second * time.Duration(t))
	// 		continue
	// 	}
	//
	// 	var foreground = vxconf.GetBoolR("server.foreground")
	// 	if noHeartbeatLog == false && foreground {
	// 		// 心跳信息不要记录到日志中，因此这里不使用 logger
	// 		logrus.Debugf("HEARTBEAT: etcd service heartbeat ok. per %d seconds.", ttl+1)
	// 	}
	// 	time.Sleep(time.Second * time.Duration(ttl))
	// }
	//
	// logrus.Debugf("etcd heartbeat loop stopped, or exiting...")
}
