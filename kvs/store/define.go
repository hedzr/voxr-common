/*
 * Copyright © 2019 Hedzr Yeh.
 */

package store

import "time"

const (
	DEFAULT_CONSUL_HOST      = "consul.ops.local"
	DEFAULT_CONSUL_LOCALHOST = "localhost"
	DEFAULT_CONSUL_PORT      = 8500
	DEFAULT_CONSUL_SCHEME    = "http"

	SERVICE_CONSUL_API = "consulapi"
	SERVICE_DB         = "test-rds"
	SERVICE_MQ         = "test-mq"
	SERVICE_CACHE      = "test-redis"

	KEY_WAS_SETUP   = "ops/config/common"
	VALUE_WAS_SETUP = "---"
)

type (
	ConsulConfig struct {
		Scheme                         string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
		Addr                           string `json:"addr" yaml:"addr"`
		Insecure                       bool   `json:"insecure,omitempty" yaml:"insecure,omitempty"`
		CertFile                       string `json:"cert,omitempty" yaml:"cert,omitempty"`
		KeyFile                        string `json:"key,omitempty" yaml:"key,omitempty"`
		CACertFile                     string `json:"cacert,omitempty" yaml:"cacert,omitempty"`
		Username                       string `json:"username,omitempty" yaml:"username,omitempty"`
		Password                       string `json:"password,omitempty" yaml:"password,omitempty"`
		Root                           string `json:"root,omitempty" yaml:"root,omitempty"`
		DeregisterCriticalServiceAfter string `json:"deregister-critical-service-after,omitempty" yaml:"deregister-critical-service-after,omitempty"`
	}

	// Etcdtool configuration struct.
	Etcdtool struct {
		Peers            string        `json:"peers,omitempty" yaml:"peers,omitempty" toml:"peers,omitempty"`
		Cert             string        `json:"cert,omitempty" yaml:"cert,omitempty" toml:"cert,omitempty"`
		Key              string        `json:"key,omitempty" yaml:"key,omitempty" toml:"key,omitempty"`
		CA               string        `json:"ca,omitempty" yaml:"ca,omitempty" toml:"peers,omitempty"`
		User             string        `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
		Timeout          time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
		CommandTimeout   time.Duration `json:"command-timeout,omitempty" yaml:"command-timeout,omitempty" toml:"command-timeout,omitempty"`
		Routes           []Route       `json:"routes" yaml:"routes" toml:"routes"`
		PasswordFilePath string        `json:"-,omitempty" yaml:",omitempty" toml:",omitempty"`
		Root             string        `json:"root,omitempty" yaml:"root,omitempty" toml:"root,omitempty"`
		NoHeartbeatLog   bool          `json:"no-heartbeat-log,omitempty" yaml:"no-heartbeat-log,omitempty" toml:"noHeartbeatLog,omitempty"`
	}

	// Route configuration struct.
	Route struct {
		Regexp string `json:"regexp" yaml:"regexp" toml:"regexp"`
		Schema string `json:"schema" yaml:"schema" toml:"schema"`
	}
)

var (
	DefaultConsulConfig = ConsulConfig{
		Scheme:   DEFAULT_CONSUL_SCHEME,
		Addr:     "127.0.0.1:8500",
		Insecure: true,
	}
)

// K/V Layer

// GetDeregisterCriticalServiceAfter default is '30s'
//
// In Consul 0.7 and later, checks that are associated with a service
// may also contain this optional DeregisterCriticalServiceAfter field,
// which is a timeout in the same Go time format as Interval and TTL. If
// a check is in the critical state for more than this configured value,
// then its associated service (and all of its associated checks) will
// automatically be deregistered.
func (s *ConsulConfig) GetDeregisterCriticalServiceAfter() string {
	if len(s.DeregisterCriticalServiceAfter) == 0 {
		return "30s"
	}
	return s.DeregisterCriticalServiceAfter
}
