/*
 * Copyright © 2019 Hedzr Yeh.
 */

package stats

import (
	"github.com/hedzr/cmdr/conf"
	"net/http"
	"strconv"
	"sync"
	"time"

	// influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"

	"fmt"
	"github.com/hedzr/voxr-common/xs/tsdb"
	"github.com/labstack/echo"
	"strings"
)

type (
	RequestStatsRecord struct {
		Time   time.Time `json:"time"`
		RealIP string    `json:"realIp"`
	}
	RequestStats struct {
		Count    int                   `json:"count"`
		Statuses map[string]int        `json:"statuses"`
		Records  []*RequestStatsRecord `json:"records,omitempty"`
	}
	Stats struct {
		Uptime       time.Time                `json:"uptime"`
		RequestCount uint64                   `json:"requestCount"`
		Statuses     map[string]int           `json:"statuses"`
		Requests     map[string]*RequestStats `json:"requests"`
		mutex        sync.RWMutex
		tsdbClient   tsdb.TinyClient
	}
)

func NewStats() *Stats {
	return &Stats{
		Uptime:     time.Now(),
		Statuses:   map[string]int{},
		Requests:   map[string]*RequestStats{},
		tsdbClient: tsdb.New(),
	}
}

func appendHeaders(tags map[string]string, req *http.Request) {
	for k, v := range req.Header {
		if strings.EqualFold(k, "X-Real-Ip") ||
			strings.EqualFold(k, "X-Forwarded-For") ||
			// strings.EqualFold(k, "Authorization") ||
			strings.EqualFold(k, "Accept") {
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

func (s *Stats) StatsIncreaseError(fwdrID string, ctx echo.Context, res *echo.Response) {

	if s.tsdbClient == nil {
		return
	}

	// timeBegin := ctx.Get("timeBegin").(time.Time)
	// timeElapsed := time.Now().Sub(timeBegin)

	// Create a point and add to batch
	tags := map[string]string{
		"instance":  conf.ServerID,
		"action":    tsdb.ACTION_INVOKE,
		"method":    ctx.Request().Method,
		"url":       ctx.Path(),
		"status":    strconv.Itoa(res.Status),
		"forwarder": fwdrID,
		// "lb":             wrr,
		// "backend_url":    backend.GetURLString(),
		// "backend_id":     backend.GetID(),
		// "backend_weight": backend.GetWeight(),
		// "backend_tag":    backend.GetTag(),
		"req_ip": ctx.Request().RemoteAddr,
		// "req_uri":        ctx.Request.RequestURI,
		// "req_ua":         res.Request.UserAgent(),
		"req_referrer": ctx.Request().Referer(),
		"real_ip":      ctx.RealIP(),
	}
	appendHeaders(tags, ctx.Request())
	fields := map[string]interface{}{
		"count": 1, // the payload
		// "time_begin":   timeBegin.UnixNano(),
		// "time_elapsed": timeElapsed.Nanoseconds(),
		"in_size":  ctx.Request().ContentLength + headerSize(ctx.Request().Header) + 4,
		"out_size": res.Size, // + headerSize(res.Header) + 4,
		// "size":           strconv.FormatInt(res.ContentLength, 10),
		// "req_size":       strconv.FormatInt(res.Request.ContentLength, 10),
	}

	s.tsdbClient.Add(tags, fields)
}

// Process is the middleware function.
func (s *Stats) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		logrus.Debugf("[global-stats][Process] IN: [%s] %s", c.Request().Method, c.Path())
		if err = next(c); err != nil {
			// c.Error(err)

			logrus.Error("[stats] Process: next(c) failed.")
			logrus.Errorf("      - Error: %v, Method: %v, req: %v", err, c.Request().Method, c.Path())

			fwdr := c.Get("fwdr")
			sss := c.Response().Status
			msID := ""
			fwdrID := ""
			if !(fwdr != nil && fwdr.(bool)) {
				fid := c.Get("fwdrID")
				if fid != nil {
					fwdrID = fid.(string)
				}
				mid := c.Get("msID")
				if mid != nil {
					msID = mid.(string)
				}

				// ms := c.Get("ms")
				// if ms == nil || len(ms.(string)) == 0 {
				err = &echo.HTTPError{Code: http.StatusNotFound, Message: err.Error()}
				c.Response().Status = http.StatusNotFound
				sss := c.Response().Status

				// save forwarder failure record to TSDB: 404, 5xx, ...
				// if sss >= 300 || msID == nil || len(msID.(string)) == 0 {
				logrus.Debugf("      - status: %d, fwdr-id: %s, ms-id: %s, c: %v", sss, fwdrID, msID, c.Path())
				// }
			} else {
				// save others entry failure records: /login, /signin, metrics, stats, ...
				// and save static entry failure records too.
				// if sss >= 300 || msID == nil || len(msID.(string)) == 0 {
				logrus.Debugf("      - status: %d, c: %v", sss, c.Path())
				// }
			}
			s.StatsIncreaseError(fwdrID, c, c.Response())
		}

		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.RequestCount++

		status := strconv.Itoa(c.Response().Status)
		s.Statuses[status]++

		uri := c.Path()
		if s.Requests[uri] == nil {
			s.Requests[uri] = &RequestStats{
				Count:    0,
				Statuses: map[string]int{},
				Records:  []*RequestStatsRecord{},
			}
		}

		s.Requests[uri].Count++
		s.Requests[uri].Statuses[status]++

		// 访问记录应该用TSDB，这里只保留任意uri被请求的计数, 为了削减内存占用而放弃其他记录
		// 同时，这里的访问计数也在api调用中被返回。
		// s.Requests[uri].Records = append(s.Requests[uri].Records, RequestStatsRecord{
		// 	Time: time.Now(),
		// 	RealIP: c.RealIP(),
		// })

		logrus.Debugf("[global-stats][Process] IN: [%s] %s | DONE |", c.Request().Method, c.Path())
		return
	}
}

// Handle is the endpoint to get stats.
func (s *Stats) Handle(c echo.Context) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return c.JSON(http.StatusOK, s)
}

// ServerHeader middleware adds a `Server` header to the response.
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set(echo.HeaderServer, conf.ServerTag)
		return next(ctx)
	}
}
