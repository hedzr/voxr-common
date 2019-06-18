/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"strconv"
)

func HashPut(key, field string, val interface{}) {
	st := hub.rw.HSet(key, field, val)
	if st.Err() != nil {
		logrus.Fatalf("HashPut(%v, %v, %v) 坏了，redis写不了了: %v", key, field, val, st)
	}
}

func HashDel(key, field string) {
	if hub.rw == nil {
		return
	}

	st := hub.rw.HDel(key, field)
	if st.Err() != nil {
		logrus.Fatalf("HashGet(%v, %v) 坏了，redis写不了了: %v", key, field, st)
	}
	return
}

func HashGet(key, field string) (st *redis.StringCmd) {
	st = hub.rd.HGet(key, field)
	if st.Err() != nil {
		logrus.Fatalf("HashGet(%v, %v) 坏了，redis写不了了: %v", key, field, st)
	}
	return
}

func HashExists(key, field string) bool {
	st := hub.rd.HExists(key, field)
	if st.Err() != nil {
		logrus.Fatalf("HashExists(%v, %v) 坏了，redis写不了了: %v", key, field, st)
	}
	return st.Val()
}

func ListPut(key string, val interface{}) {
	st := hub.rw.RPush(key, val)
	if st.Err() != nil {
		logrus.Fatalf("ListPut(%v, %v) 坏了，redis写不了了: %v", key, val, st)
	}
}

func ListGet(key string) (list []interface{}) {
	return
}

func Put(key string, val interface{}) {
	st := hub.rw.Set(key, val, 0)
	if st.Err() != nil {
		logrus.Fatalf("Put(%v, %v) 坏了，redis写不了了: %v", key, val, st)
	}
}

func Get(key string) (val string) {
	v, err := hub.rd.Get(key).Result()
	if err != nil {
		logrus.Fatalf("Get(%v) 坏了，redis取不到了: %v", key, err)
	}
	val = v
	return
}

func Exists(key string) (yes bool) {
	return
}

func int2str(i int) string {
	return strconv.Itoa(i)
}

func GetRedisRw() redis.Cmdable {
	return hub.rw
}

func GetRedisRd() redis.Cmdable {
	return hub.rd
}

func GetClient() redis.Cmdable {
	return hub.rd
}

func GetClientForWrite() redis.Cmdable {
	return hub.rw
}

func GetClientFor(readOnly bool) redis.Cmdable {
	if readOnly {
		return hub.rd
	} else {
		return hub.rw
	}
}

func Do(args ...interface{}) (cmd *redis.Cmd) {
	if cc, ok := hub.rw.(*redis.ClusterClient); ok {
		cmd = cc.Do(args...)
	} else if bc, ok := hub.rw.(*redis.Client); ok {
		cmd = bc.Do(args...)
	}
	return
}

func Watch(fn func(*redis.Tx) error, keys ...string) error {
	if cc, ok := hub.rw.(*redis.ClusterClient); ok {
		return cc.Watch(fn, keys...)
	} else if bc, ok := hub.rw.(*redis.Client); ok {
		return bc.Watch(fn, keys...)
	}
	return nil
}

func PipelineExec(fn func(pipeline redis.Pipeliner) (err error)) (err error) {
	if cc, ok := hub.rw.(*redis.ClusterClient); ok {
		pl := cc.Pipeline()
		err = fn(pl)
		pl.Exec()
	} else if bc, ok := hub.rw.(*redis.Client); ok {
		pl := bc.Pipeline()
		err = fn(pl)
		pl.Exec()
	}
	return nil
}

func TxPipelineExec(fn func(pipeline redis.Pipeliner) (err error)) (err error) {
	if cc, ok := hub.rw.(*redis.ClusterClient); ok {
		pl := cc.TxPipeline()
		err = fn(pl)
		pl.Exec()
	} else if bc, ok := hub.rw.(*redis.Client); ok {
		pl := bc.TxPipeline()
		err = fn(pl)
		pl.Exec()
	}
	return nil
}

func Publish(channel string, message interface{}) *redis.IntCmd {
	return hub.rw.Publish(channel, message)
}

func Subscribe(channels ...string) (pubsub *redis.PubSub) {
	if cc, ok := hub.rw.(*redis.ClusterClient); ok {
		pubsub = cc.Subscribe(channels...)
	} else if bc, ok := hub.rw.(*redis.Client); ok {
		pubsub = bc.Subscribe(channels...)
	}
	return
}
