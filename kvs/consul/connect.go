/*
 * Copyright © 2019 Hedzr Yeh.
 */

package consul

import (
	"context"
	"github.com/hedzr/voxr-common/kvs/consul/consul_util"
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/sirupsen/logrus"
	"time"
)

func New(e *store.ConsulConfig) store.KVStore {
	cli, _, err := consul_util.GetConsulConnection(e)
	if err != nil {
		logrus.Fatalf("FATAL: CANNOT CONNECT TO CONSUL. %v", err)
		return nil
	}

	store := KVStoreConsul{
		cli,
		ROOT_KEY,
		e,
		nil,
		nil,
		5 * time.Second,
		make(chan bool),
		make(chan bool, 3),
		false,
		"",
	}

	store.ctx, store.cancelFunc = contextWithCommandTimeout(store.Timeout)
	return &store
}

func contextWithCommandTimeout(commandTimeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), commandTimeout)
}
