/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"github.com/hedzr/voxr-common/kvs/store"
	"github.com/hedzr/voxr-common/xs/proxy"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"net/url"
)

const (
	TYPE_MS   = "ms"
	TYPE_HTTP = "http"
	TYPE_MOCK = "mock"

	AUTH_TYPE_NONE   = "none"
	AUTH_TYPE_BASIC  = "basic"
	AUTH_TYPE_HEADER = "header"
	AUTH_TYPE_JWT    = "jwt"
	AUTH_TYPE_OAUTH1 = "oauth1"
	AUTH_TYPE_OAUTH2 = "oauth2"

	DEFAULT_SERVICE_SUFFIX = ".service.consul."
	DEFAULT_DNS_SERVER     = "127.0.0.1:8600"
)

type (
	// Auth 提供目标服务的认证方案，注意通常不需要
	Auth struct {
		// basic auth, header(jwt), key-based: oauth1, oauth2, ...
		Type      string `json:"type,omitempty" yaml:"type,omitempty"`
		Username  string `json:"userid,omitempty" yaml:"userid,omitempty"`
		Password  string `json:"passwd,omitempty" yaml:"passwd,omitempty"`
		Header    string `json:"header,omitempty" yaml:"header,omitempty"`
		AppKey    string `json:"appkey,omitempty" yaml:"appkey,omitempty"`
		AppSecret string `json:"secret,omitempty" yaml:"secret,omitempty"`
	}

	// To presents where to go to a micro-service, includes its well-known service name in registry
	To struct {
		MS      string `json:"ms,omitempty" yaml:"ms,omitempty"`
		Context string `json:"context,omitempty" yaml:"context,omitempty"` // ContextUrl前缀
	}

	// Targets present a set of endpoints about that external/internal RESTful service
	Targets struct {
		Value []string `json:"target,omitempty" yaml:"target,omitempty"`
	}

	Mock struct {
		Methods []string    `json:"methods,flow,omitempty" yaml:"methods,flow,omitempty"`
		Object  interface{} `json:"object,omitempty" yaml:"object,flow,omitempty"`
		Text    string      `json:"text,omitempty" yaml:"text,omitempty"`
	}

	CircuitBreakAt struct {
		Failed        int `json:"failed,omitempty" yaml:"failed,omitempty"`
		Timeout       int `json:"timeout,omitempty" yaml:"timeout,omitempty"`
		CannotResolve int `json:"cannotresolve,omitempty" yaml:"cannotresolve,omitempty"`
		HTTP5xx       int `json:"5xx,omitempty" yaml:"5xx,omitempty"`
		HTTP4xx       int `json:"4xx,omitempty" yaml:"4xx,omitempty"`
	}

	LoadBalanceSmart struct {
		URLHash  []string `json:"url-hash,omitempty" yaml:"url-hash,omitempty"`
		IPHash   []string `json:"ip-hash,omitempty" yaml:"ip-hash,omitempty"`
		Versions []string `json:"ver,omitempty" yaml:"ver,omitempty"`
		Weight   []int    `json:"weight,omitempty" yaml:"weight,omitempty"`
		// verConstraints []*lb.VersionConstraintWeight
	}

	LoadBal struct {
		ID             string            `json:"id,omitempty" yaml:"id,omitempty"`
		Algorithm      string            `json:"alg" yaml:"alg"`
		CircuitBreakAt *CircuitBreakAt   `json:"break" yaml:"break"`
		Smart          *LoadBalanceSmart `json:"smart,omitempty" yaml:"smart,omitempty"`
	}

	// RequestStatsRecord struct {
	// 	Time   time.Time `json:"time"`
	// 	RealIP string    `json:"realIp"`
	// }
	//
	// RequestStats struct {
	// 	Count    int                   `json:"count"`
	// 	Statuses map[string]int        `json:"statuses"`
	// 	Records  []*RequestStatsRecord `json:"records,omitempty"`
	// }
	//
	// Stats struct {
	// 	Uptime       time.Time                `json:"uptime"`
	// 	RequestCount uint64                   `json:"requestCount"`
	// 	Statuses     map[string]int           `json:"statuses"`
	// 	Requests     map[string]*RequestStats `json:"requests"`
	// 	mutex        sync.RWMutex
	// }

	// Forwarder to define a forwarder/proxy/reverse-proxy to micro-service/external-service/...
	FwdItem struct {
		ID               string   `json:"id" yaml:"id"`
		Type             string   `json:"type" yaml:"type"`
		Match            string   `json:"match" yaml:"match"`
		To               To       `json:"to,omitempty" yaml:"to,omitempty"`
		Targets          []string `json:"targets,flow,omitempty" yaml:"targets,flow,omitempty"`
		Mocks            []*Mock  `json:"mocks,omitempty" yaml:"mocks,omitempty"`
		Lb               *LoadBal `json:"lb,omitempty" yaml:"lb,omitempty"`
		DowngradeToHttp1 bool     `json:"downgrade-to-http1,omitempty" yaml:"downgrade-to-http1,omitempty"`
		ReverseRewrite   bool     `json:"reverse-rewrite,omitempty" yaml:"reverse-rewrite,omitempty"`
		Insecure         bool     `json:"insecure,omitempty" yaml:"insecure,omitempty"`
		NoTrailingSlash  bool     `json:"no-trailing-slash,omitempty" yaml:"no-trailing-slash,omitempty"`
		Description      string   `json:"desc,omitempty" yaml:"desc,omitempty"`
		WhiteLists       []string `json:"white-lists,omitempty" yaml:"white-lists,omitempty"`
		BlackLists       []string `json:"black-lists,omitempty" yaml:"black-lists,omitempty"`
		Disabled         bool     `json:"disabled,omitempty" yaml:"disabled,omitempty"`
		BreakAtFailured  int      `json:"break-at-failure,omitempty" yaml:"break-at-failure,omitempty"`
		Auth             Auth     `json:"auth,omitempty" yaml:"auth,omitempty"`
		ForceUpdate      bool     `json:"force-update,omitempty" yaml:"force-update,omitempty"`

		proxy       *proxy.TinyProxy
		balancer    middleware.ProxyBalancer
		handlerFunc echo.MiddlewareFunc

		// tsdbClient tsdb.TinyClient
		// proxyConfig      *middleware.ProxyConfig

		exitCh chan struct{}
	}

	Registrar struct {
		Enabled    bool                           `json:"enabled" yaml:"enabled"`
		Source     string                         `json:"source" yaml:"source"`
		Env        string                         `json:"env" yaml:"env"`
		TTL        int64                          `json:"ttl,omitempty" yaml:"ttl,omitempty"`
		Consul     map[string]*store.ConsulConfig `json:"consul,omitempty" yaml:"consul,omitempty"`
		Etcd       map[string]*store.Etcdtool     `json:"etcd,omitempty" yaml:"etcd,omitempty"`
		DNSAtFirst bool                           `json:"dns-at-first,omitempty" yaml:"dns-at-first,omitempty"`
		store      store.KVStore                  `json:"_,omitempty" yaml:",omitempty"`
	}

	// Stud is just a structure stub used for yaml converter
	Stub struct {
		Forwarders []FwdItem `json:"forwarders" yaml:"forwarders,flow"`
	}

	// Config a forwarders configurations inside main app config file (meta.yaml)
	Config struct {
		Forwarders     []*FwdItem         `json:"forwarders" yaml:"forwarders,flow"`
		Registrar      Registrar          `json:"registrar" yaml:"registrar,flow"`
		Skipper        middleware.Skipper `json:"_,omitempty" yaml:",omitempty"` // Skipper defines a function to skip middleware.
		UniqueFwdIDMap map[string]int     `json:"_,omitempty" yaml:",omitempty"`
		UniqueLbIDMap  map[string]int     `json:"_,omitempty" yaml:",omitempty"`
		UniqueLbIDInt  int                `json:"_,omitempty" yaml:",omitempty"`
		Echo           *echo.Echo         `json:"_,omitempty" yaml:",omitempty"`
	}

	VersionSetter interface {
		PutVersion(res *http.Response, target *url.URL, name string, version string)
	}
)
