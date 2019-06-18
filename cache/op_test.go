/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache_test

import (
	"fmt"
	"github.com/go-redis/redis"
	"testing"
)

const (
	// addr    = "127.0.0.1:26379"
	// cluster = false
	addr    = "127.0.0.1:6379"
	cluster = true
)

func setup() func() {
	StartWithConfig(&Config{
		Peers:         []string{addr},
		EnableCluster: cluster,
	})
	return func() {
		Stop()
	}
}

func TestConn(t *testing.T) {
	defer setup()()

	Put("r:v:o:k", 1)
}

func TestPublish(t *testing.T) {
	defer setup()()
	done := make(chan struct{})
	Publish("mychannel", "hello budy!\n")
	go func() {
		pubsub := Subscribe("mychannel")
		msg, _ := pubsub.Receive()
		fmt.Println("Receive from channel:", msg)
		done <- struct{}{}
	}()

	<-done
}

func TestPipeline(t *testing.T) {
	defer setup()()
	PipelineExec(func(pl redis.Pipeliner) (err error) {
		pl.Set("pipe", 0, 0)
		pl.Incr("pipe")
		pl.Incr("pipe")
		pl.Incr("pipe")
	})
}
