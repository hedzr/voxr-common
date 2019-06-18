/*
 * Copyright © 2019 Hedzr Yeh.
 */

package etcd

import (
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

func watchRunner(rch clientv3.WatchChan, fn store.WatchFunc, stopCh chan bool) {
	for {
		select {
		case wresp := <-rch:
			for _, ev := range wresp.Events {
				// fmt.Printf("watch event - %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				// t := int(ev.Type)
				// et := store.Event_EventType(ev.Type)
				fn(store.Event_EventType(ev.Type), ev.Kv.Key, ev.Kv.Value)
			}

		case ok := <-stopCh:
			logrus.Debugf("etcd watchRunner is exiting... %v", ok)
			return
		}
	}
	// close(rch)
}
