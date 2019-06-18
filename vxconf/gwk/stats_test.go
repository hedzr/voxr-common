/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk_test

import (
	"testing"
	"time"
)

func TestRateLimit1(t *testing.T) {

}

// see also:
//   https://github.com/golang/go/wiki/RateLimiting
//   https://stackoverflow.com/questions/27187617/how-would-i-limit-upload-and-download-speed-from-the-server-in-golang
// limit http get rate:
//   https://stackoverflow.com/questions/27869858/limiting-bandwidth-of-http-get
//

func callApi(req interface{}) {
	// client.Call("Service.Method", req, ...)
}

// func a1() {
// 	rate := time.Second / 10
// 	throttle := time.Tick(rate)
// 	for req := range requests {
// 		<-throttle // rate limit our Service.Method RPCs
// 		go callApi(req)
// 	}
// }

func burstA1() {
	var requests = []string{"url1"}
	rate := time.Second / 10
	burstLimit := 100
	tick := time.NewTicker(rate)
	throttle := make(chan time.Time, burstLimit)
	go func() {
		defer tick.Stop()
		for t := range tick.C {
			select {
			case throttle <- t:
			default:
			}
		} // exits after tick.Stop()
	}()
	for req := range requests {
		<-throttle // rate limit our Service.Method RPCs
		go callApi(req)
	}
}
