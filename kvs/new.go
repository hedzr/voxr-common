/*
 * Copyright © 2019 Hedzr Yeh.
 */

package kvs

import (
	"github.com/hedzr/voxr-common/kvs/consul"
	"github.com/hedzr/voxr-common/kvs/etcd"
	"github.com/hedzr/voxr-common/kvs/store"
)

const (
	STORE_ETCD = iota
	STORE_CONSUL
)

func New(e interface{}) store.KVStore {
	if ee, ok := e.(*store.Etcdtool); ok {
		return NewETCD(ee)
	}
	if ee, ok := e.(*store.ConsulConfig); ok {
		return NewConsul(ee)
	}
	return nil
}

func NewETCD(e *store.Etcdtool) store.KVStore {
	return etcd.New(e)
}

func NewConsul(e *store.ConsulConfig) store.KVStore {
	return consul.New(e)
}
