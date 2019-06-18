/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"fmt"
	"github.com/hedzr/voxr-common/xs/proxy"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

// 兼容旧版本算法，旧版本使用 TinyProxy
func (s *FwdItem) SetProxy(p *proxy.TinyProxy) {
	if s.proxy != nil {
		s.proxy.Stop()
	}
	s.proxy = p
}

func (s *FwdItem) Proxy() (p *proxy.TinyProxy) {
	p = s.proxy
	return
}

func (s *FwdItem) MiddlewareFunc() (fn echo.MiddlewareFunc) {
	fn = s.handlerFunc
	return
}

func (s *FwdItem) SetMiddlewareFunc(fn echo.MiddlewareFunc) {
	s.handlerFunc = fn
}

func (s *FwdItem) Balancer() (bal middleware.ProxyBalancer) {
	bal = s.balancer
	return
}

func (s *FwdItem) SetBalancer(bal middleware.ProxyBalancer) {
	s.balancer = bal
}

func (s *FwdItem) BalancerIsEmpty() bool {
	return s.balancer == nil
}

func (s *FwdItem) BuildUrl(ip net.IP, port uint16, req *http.Request) *url.URL {
	path0 := req.URL.Path
	path1 := path0[len(s.Match):]
	host := ip.String()
	if ip.To4() == nil { // ipv6
		host = fmt.Sprintf("[%s]", ip.String())
	}
	url1 := fmt.Sprintf("http://%s:%d%s%s", host, port, s.To.Context, path1)
	//
	// url1 := fmt.Sprintf("http://%s:%d%s%s", host, port, forwarder.To.Context, "/")
	url, err := url.Parse(url1)
	url.ForceQuery = req.URL.ForceQuery
	url.RawQuery = req.URL.RawQuery
	url.Fragment = req.URL.Fragment
	if err != nil {
		log.Errorf("cannot parse url string of service '%s'. the url is '%s'.", s.To.MS, url1)
		return nil
	}
	return url
}

// PutVersion 专用于放置从后端响应头中提取到的版本号相关信息
func (s *FwdItem) PutVersion(res *http.Response, target *url.URL) {
	service := res.Header.Get("x-service")
	pattern, _ := regexp.Compile("^([^/]+)/([0-9.]+)")
	r := pattern.FindStringSubmatchIndex(service)
	if len(r) >= 6 && s.balancer != nil {
		// log.Debug("version pattern found: ", r)
		name := service[r[2]:r[3]]
		version := service[r[4]:r[5]]
		// log.Debugf("        name = %s, version = %s, url = %s", name, version, res.Request.URL.String())
		var vs VersionSetter = s.balancer.(VersionSetter)
		vs.PutVersion(res, target, name, version)
	}
}

func (s *FwdItem) buildLbIDs() {
	if s.Lb != nil && s.Lb.Smart != nil {
		if len(s.Lb.ID) == 0 {
			// generate an unique id for load balancer
			ThisConfig.UniqueLbIDInt++
			// 暂时使用固定算法，不使用 UniqueLbIDInt
			s.Lb.ID = fmt.Sprintf("%s-lb", s.ID)
		}
	}
}

func (s *FwdItem) Shutdown() {
	// log.Infof("   forwarder shutting down.")
	if s.exitCh != nil {
		// s.exitCh <- struct{}{}
	}
	// if s.tsdbClient != nil {
	// 	s.tsdbClient.Shutdown()
	// 	s.tsdbClient = nil
	// }
}

func (s *FwdItem) Init() {
	if s.exitCh == nil {
		s.exitCh = make(chan struct{})
	}
	// if s.tsdbClient == nil {
	// 	s.tsdbClient = tsdb.New()
	// }
	s.buildLbIDs()
}
