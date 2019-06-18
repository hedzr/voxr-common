/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"github.com/go-redis/redis"
	"strings"
	"time"
)

type (
	Hub struct {
		clients map[redis.Cmdable]bool
		rw      redis.Cmdable
		rd      redis.Cmdable
		exitCh  chan bool
		exited  bool
		// broadcast  chan []byte
		// register   chan *redis.ClusterClient
		// unregister chan *redis.ClusterClient
		reading chan *kvpair
		writing chan *kvpair
	}

	kvpair struct {
		key, value string
	}

	Config struct {
		Peers         []string      `yaml:"peers"`
		Username      string        `yaml:"user"'`
		Password      string        `yaml:"pass"`
		Db            int           `yaml:"db"`
		ReadonlyRoute bool          `yaml:readonly-route`
		DialTimeout   time.Duration `yaml:"dial-timeout"`
		ReadTimeout   time.Duration `yaml:"read-timeout"`
		WriteTimeout  time.Duration `yaml:"write-timeout"`
		EnableCluster bool          `yaml:"enable-cluster"`
	}
)

var (
	hub = Hub{
		clients: make(map[redis.Cmdable]bool),
		rw:      nil,
		rd:      nil,
		exitCh:  make(chan bool),
		exited:  true,
		// broadcast:  make(chan []byte),
		// register:   make(chan *redis.ClusterClient),
		// unregister: make(chan *redis.ClusterClient),
		reading: make(chan *kvpair),
		writing: make(chan *kvpair),
	}
)

func (h *Hub) start(config *Config) {

	JwtInit()

	// for debugging only, to simulate inserting automatically
	// go h.run()

	if config.DialTimeout == 0 {
		config.DialTimeout = 10 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 20 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 30 * time.Second
	}

	if config.EnableCluster {
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          config.Peers,
			RouteByLatency: config.ReadonlyRoute,
			DialTimeout:    config.DialTimeout,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			Password:       config.Password,
		})
		h.clients[client] = true
		h.rd = client

		clientW := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Peers,
			ReadOnly:     false,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Password:     config.Password,
		})
		h.clients[clientW] = true
		h.rw = clientW
	} else {
		// peers := vxconf.GetStringSliceR("server.pub.deps.redis.peers", nil)
		client := redis.NewClient(&redis.Options{
			Addr:         strings.Join(config.Peers, ","),
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Password:     config.Password,
		})
		h.clients[client] = true
		h.rd = client

		clientW := redis.NewClient(&redis.Options{
			Addr:         strings.Join(config.Peers, ","),
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			Password:     config.Password,
		})
		h.clients[clientW] = true
		h.rw = clientW
	}

}

func (h *Hub) stop() {
	for k := range hub.clients {
		if cc, ok := k.(*redis.ClusterClient); ok {
			cc.Close()
		} else if bc, ok := k.(*redis.Client); ok {
			bc.Close() // it will break the for loop in ws_hello()
		}
	}
	h.clients = make(map[redis.Cmdable]bool)
	h.rw = nil
	h.rd = nil
	if !h.exited {
		h.exitCh <- true
	}
}
