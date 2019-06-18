/*
 * Copyright © 2019 Hedzr Yeh.
 */

package proxy

import (
	"flag"
	"fmt"
	"github.com/hedzr/cmdr/conf"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

type TinyProxy struct {
	target        *url.URL
	proxy         *httputil.ReverseProxy
	routePatterns []*regexp.Regexp // add some route patterns with regexp
}

func New(target string) *TinyProxy {
	url, _ := url.Parse(target)

	return &TinyProxy{target: url, proxy: httputil.NewSingleHostReverseProxy(url)}
}

func (p *TinyProxy) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Proxy", conf.ServerTag)

	if p.routePatterns == nil || p.parseWhiteList(r) {
		p.proxy.ServeHTTP(w, r)
	}
}

func (p *TinyProxy) Stop() {
	// TODO 确认TinyProxy并不需要一个停止服务的调用吗？
}

func (p *TinyProxy) StandardHandle(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

func (p *TinyProxy) UpdateTarget(url *url.URL) {
	p.target = url
	p.proxy = httputil.NewSingleHostReverseProxy(url)
}

func (p *TinyProxy) parseWhiteList(r *http.Request) bool {
	for _, regexp := range p.routePatterns {
		fmt.Println(r.URL.Path)
		if regexp.MatchString(r.URL.Path) {
			// let's forward it
			return true
		}
	}
	fmt.Printf("Not accepted routes %x", r.URL.Path)
	fmt.Println()
	return false
}

func mainTest() {
	const (
		defaultPort             = ":80"
		defaultPortUsage        = "default server port, ':80', ':8080'..."
		defaultTarget           = "http://127.0.0.1:8080"
		defaultTargetUsage      = "default redirect url, 'http://127.0.0.1:8080'"
		defaultWhiteRoutes      = `^\/$|[\w|/]*.js|/path|/path2`
		defaultWhiteRoutesUsage = "list of white route as regexp, '/path1*,/path2*...."
	)

	// flags
	port := flag.String("port", defaultPort, defaultPortUsage)
	url := flag.String("url", defaultTarget, defaultTargetUsage)
	routesRegexp := flag.String("routes", defaultWhiteRoutes, defaultWhiteRoutesUsage)

	flag.Parse()

	fmt.Printf("server will run on : %s\n", *port)
	fmt.Printf("redirecting to :%s\n", *url)
	fmt.Printf("accepted routes :%s\n", *routesRegexp)

	//
	reg, _ := regexp.Compile(*routesRegexp)
	routes := []*regexp.Regexp{reg}

	// proxy
	proxy := New(*url)
	proxy.routePatterns = routes

	// server
	http.HandleFunc("/", proxy.handle)
	http.ListenAndServe(*port, nil)
}
