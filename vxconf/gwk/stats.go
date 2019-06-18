/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"github.com/hedzr/voxr-common/kvs/id"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	// influxdb "github.com/influxdata/influxdb/client/v2"
	// log "github.com/sirupsen/logrus"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	GwBackend interface {
		GetID() string
		GetURLString() string
		GetTag() string
		GetWeight() string
	}

	defaultBackend struct{}

	GwBalancer interface {
		GetID() string
		// GetBackend returns the real backend object from the balancer
		GetBackend(target *middleware.ProxyTarget) GwBackend
	}
)

var (
	DefaultGwBackend = &defaultBackend{}
)

func (s *defaultBackend) GetID() string        { return "" }
func (s *defaultBackend) GetURLString() string { return "" }
func (s *defaultBackend) GetTag() string       { return "" }
func (s *defaultBackend) GetWeight() string    { return "" }

func (s *FwdItem) StatsIncreaseHit(target *middleware.ProxyTarget, ctx echo.Context, res *http.Request) {
	// do nothing.
	// req.URL is real external request url
	// log.Debugf(">> StatsIncreaseHit: target=%s", target.URL.String())
	ctx.Set("msHit", time.Now())
	ctx.Set("msID", s.To.MS)
	ctx.Set("fwdrID", s.ID)
	ctx.Set("fwdr", true)
}

func (s *FwdItem) StatsIncreaseReq(target *middleware.ProxyTarget, ctx echo.Context, req *http.Request, url *url.URL, realIP string) {
	// log.Debugf(">> StatsIncreaseReq: target=%s, url=%s", target.URL.String(), url.String())
	ctx.Set("timeBegin", time.Now())
	// c := s.ensureTsdbClient()
	// if c == nil { return }
	//
	// dbName := vxconf.GetStringR("server.tsdb.current.database")
	//
	// // Create a new point batch
	// bp, _ := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
	// 	Database:  dbName,
	// 	Precision: "ns",
	// })
	//
	// var backend GwBackend = DefaultGwBackend
	// if bal, ok := s.balancer.(GwBalancer); ok {
	// 	backend = bal.GetBackend(target)
	// }
	//
	// // Create a point and add to batch
	// tags := map[string]string{
	// 	"instance":       gwk.ServerID,
	// 	"action":         ACTION_INVOKE,
	// 	"method":         req.Method,
	// 	"url":            url.String(), //req.URL.String(), // 可能是修正过后的URL了，see #ProxyMS()
	// 	"status":         "",
	// 	"size":           "0",
	// 	"backend_url":    backend.GetURLString(),
	// 	"backend_id":     backend.GetID(),
	// 	"backend_weight": backend.GetWeight(),
	// 	"backend_tag":    backend.GetTag(),
	// 	"real_ip":        realIP,
	// 	"req_ip":         req.RemoteAddr,
	// 	"req_uri":        req.RequestURI,
	// 	"req_size":       strconv.FormatInt(req.ContentLength, 10),
	// 	"req_ua":         req.UserAgent(),
	// 	"req_referrer":   req.Referer(),
	// }
	// s.appendHeaders(tags, req)
	// fields := map[string]interface{}{
	// 	"count":  1, // the payload
	// 	"system": 53.3,
	// 	"user":   46.6,
	// }
	// pt, err := influxdb.NewPoint("stats", tags, fields, time.Now())
	// if err != nil {
	// 	log.Errorf("[TSDB] NewPoint failed. Error: %v", err)
	// }
	// bp.AddPoint(pt)
	//
	// // Write the batch
	// if err := c.Write(bp); err != nil {
	// 	log.Errorf("[TSDB] Write failed. Error: %v", err)
	// }
}

func (s *FwdItem) StatsIncreaseResp(target *middleware.ProxyTarget, ctx echo.Context, res *http.Response, url *url.URL, realIP string) {

	timeBegin := ctx.Get("timeBegin").(time.Time)
	/*timeElapsed*/ _ = time.Now().Sub(timeBegin)
	originalUrl := ctx.Get("req-url").(string)

	var backend GwBackend = DefaultGwBackend
	wrr := ""
	if s.Lb != nil && len(s.Lb.ID) > 0 {
		wrr = s.Lb.ID
	}
	if bal, ok := s.balancer.(GwBalancer); ok {
		backend = bal.GetBackend(target)
		// wrr = bal.GetID()
	}

	// log.Debugf(">> StatsIncreaseResp: target=%s, url=%s, start=%s, elapsed=%s", target.URL.String(), url.String(), timeBegin.String(), timeElapsed.String())

	// c := s.ensureTsdbClient()
	// if c == nil {
	// 	return
	// }

	// Create a point and add to batch
	tags := map[string]string{
		"instance":       id.GetInstanceId(),
		"action":         "invoke", // tsdb.ACTION_INVOKE,
		"method":         res.Request.Method,
		"url":            originalUrl,
		"status":         strconv.Itoa(res.StatusCode),
		"forwarder":      s.ID,
		"lb":             wrr,
		"backend_url":    backend.GetURLString(),
		"backend_id":     backend.GetID(),
		"backend_weight": backend.GetWeight(),
		"backend_tag":    backend.GetTag(),
		"req_ip":         res.Request.RemoteAddr,
		// "req_uri":        res.Request.RequestURI,
		// "req_ua":         res.Request.UserAgent(),
		"req_referrer": res.Request.Referer(),
		"real_ip":      realIP,
	}
	s.appendHeaders(tags, res.Request)
	// fields := map[string]interface{}{
	// 	"count":        1, // the payload
	// 	"time_begin":   timeBegin.UnixNano(),
	// 	"time_elapsed": timeElapsed.Nanoseconds(),
	// 	"in_size":      res.ContentLength + headerSize(res.Request.Header) + 4,
	// 	"out_size":     res.Request.ContentLength + headerSize(res.Header) + 4,
	// 	//"size":           strconv.FormatInt(res.ContentLength, 10),
	// 	//"req_size":       strconv.FormatInt(res.Request.ContentLength, 10),
	// }
	//
	// s.tsdbClient.Add(tags, fields)
}

func (s *FwdItem) appendHeaders(tags map[string]string, req *http.Request) {
	for k, v := range req.Header {
		if strings.EqualFold(k, "X-Real-Ip") || strings.EqualFold(k, "X-Forwarded-For") || strings.EqualFold(k, "Accept") {
			continue
		}
		tags[fmt.Sprintf("req_%s", k)] = strings.Join(v, ",")
	}
}

func headerSize(header http.Header) (l int64) {
	for k, v := range header {
		l += int64(len(k) + len(strings.Join(v, "  ")) + 2)
	}
	l += 58 // magic number
	return
}
