/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cache

import (
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func loadDefaultConfig() (config *Config) {
	config = new(Config)
	_ = vxconf.LoadSectionTo("server.pub.deps.redis", config)

	if s := os.Getenv("CACHE_ADDR"); len(s) > 0 {
		config.Peers = strings.Split(s, ",")
	}

	logrus.Debugf("      [cache] redis: %v", config.Peers)
	return
}

func Start() {
	hub.start(loadDefaultConfig())
}

func StartWithConfig(config *Config) {
	hub.start(config)
}

func Stop() {
	hub.stop()
}
