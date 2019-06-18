/*
 * Copyright © 2019 Hedzr Yeh.
 */

package store_test

import (
	"github.com/hedzr/voxr-common/kvs"
	"github.com/hedzr/voxr-common/kvs/etcd"
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/hedzr/voxr-common/tool"
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestThisHostname(t *testing.T) {
	h := tool.ThisHostname()
	t.Logf("ThisHostname = %s", h)
	ip := store.ThisHost()
	t.Logf("ThisHost = %s", ip.String())
}

func getConsulStore() store.KVStore {
	store := kvs.New(&ConsulConfig{
		Scheme:                         "http",
		Addr:                           "127.0.0.1:8500",
		Insecure:                       true,
		CertFile:                       "",
		KeyFile:                        "",
		CACertFile:                     "",
		Username:                       "",
		Password:                       "",
		Root:                           "",
		DeregisterCriticalServiceAfter: "30s",
	})
	return store
}

func getEtcdStore() store.KVStore {
	store := kvs.New(&etcd.Etcdtool{
		Peers:            "127.0.0.1:2379",
		Cert:             "",
		Key:              "",
		CA:               "",
		User:             "",
		Timeout:          time.Second * 10,
		CommandTimeout:   time.Second * 5,
		Routes:           []etcd.Route{},
		PasswordFilePath: "",
		Root:             "",
	})
	return store
}

func TestGetPut(t *testing.T) {
	store := getConsulStore()

	store.Put("x", "yz")
	t.Log(store.Get("x"))
	assert.Equal(t, store.Get("x"), "yz", "expect ['x'] == 'yz'.")
}
