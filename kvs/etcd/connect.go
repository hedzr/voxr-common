/*
 * Copyright © 2019 Hedzr Yeh.
 */

package etcd

import (
	"github.com/bgentry/speakeasy"
	"github.com/hedzr/voxr-common/kvs/store"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func demo() {
	cli, err := clientv3.New(clientv3.Config{
		// Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()
}

var noHeartbeatLog = false

func New(e *store.Etcdtool) store.KVStore {
	c := NewClient(e)
	store := KVStoreEtcd{
		c,
		e,
		nil,
		nil,
		make(chan bool), make(chan bool),
	}
	noHeartbeatLog = e.NoHeartbeatLog
	store.ctx, store.cancelFunc = contextWithCommandTimeout(e.CommandTimeout)
	return &store
}

func contextWithCommandTimeout(commandTimeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), commandTimeout)
}

func newTransport(e *store.Etcdtool) *http.Transport {
	tls := transport.TLSInfo{
		TrustedCAFile: e.CA,
		CertFile:      e.Cert,
		KeyFile:       e.Key,
	}

	timeout := 30 * time.Second
	tr, err := transport.NewTransport(tls, timeout)
	if err != nil {
		warnf("WARN: %v", err)
	}

	return tr
}

func NewClient(e *store.Etcdtool) *clientv3.Client {
	cfg := clientv3.Config{
		DialTimeout: 5 * time.Second,
		Endpoints:   strings.Split(e.Peers, ","),
		// Transport:               newTransport(e),
		// HeaderTimeoutPerRequest: e.Timeout,
	}

	if !E3W_MODE {
		if !strings.HasPrefix(e.Root, "/") {
			e.Root = "/" + e.Root
		}
	}
	if e.User != "" {
		cfg.Username = e.User
		if e.PasswordFilePath != "" {
			passBytes, err := ioutil.ReadFile(e.PasswordFilePath)
			if err != nil {
				warnf("WARN: %v", err)
			}
			cfg.Password = strings.TrimRight(string(passBytes), "\n")
		} else {
			pwd, err := speakeasy.Ask("Password: ")
			if err != nil {
				warnf("WARN: %v", err)
			}
			cfg.Password = pwd
		}
	}

	cl, err := clientv3.New(cfg)
	if err != nil {
		warnf("WARN: %v", err)
	}

	return cl
}
