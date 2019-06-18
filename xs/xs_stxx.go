/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"context"
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// func quitSignal() chan os.Signal {
// 	// Wait for interrupt signal to gracefully shutdown the server with
// 	// a timeout of 10 seconds.
// 	quit := daemon.QuitSignals()
// 	return quit
// }
//
// func startAndBlock(e *echo.Echo, s *echoServerImpl) {
// 	quit := quitSignal()
//
// 	if vxconf.GetBoolR("server.foreground") {
// 		logrus.Debugf("@@ server starting, foreground.")
// 		go run(s, e, quit)
//
// 		// fmt.Println("Press CONTROL-C to exit.")
// 		<-quit
// 		fmt.Println(" caught")
// 		logrus.Debugf("@@ os interrupt signal caught, server will be shutdowning gracefully.")
// 		logrus.Debugf("@@ shutdown listener.")
// 		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
// 		defer func() {
// 			fmt.Println("    10 seconds waiting.")
// 			cancel()
// 			fmt.Println("    cancel() called.")
// 			stopCh <- true
// 			stopCh <- true
// 			// close(stopCh)
// 		}()
// 		if err := e.Shutdown(ctx); err != nil {
// 			logrus.Fatal(err)
// 		} else {
// 			fmt.Println("    e.shutdown() called.")
// 		}
// 		// <-stopCh
// 	} else {
// 		run(s, e, quit)
// 	}
// 	// fmt.Println(" end")
// }
//
// func run(s *echoServerImpl, e *echo.Echo, quit chan os.Signal) {
// 	fAddr := vxconf.GetStringR("server.address")
// 	fPort := vxconf.GetIntR("server.port")
// 	fCert := vxconf.GetStringR("server.tls.cert")
// 	fKey := vxconf.GetStringR("server.tls.key")
//
// 	if s := os.Getenv("PORT"); len(s) > 0 {
// 		if port1, err := strconv.Atoi(s); err == nil {
// 			fPort = port1
// 		}
// 	}
//
// 	addr := fAddr
// 	if strings.Index(fAddr, ":") < 0 {
// 		addr = fmt.Sprintf("%s:%d", fAddr, fPort)
// 	}
// 	f := vxconf.GetBoolR("server.foreground")
// 	if f {
// 		s.cool.OnPrintBanner(addr)
// 	} else {
// 		// daemon.Target = e
// 		// daemon.Blocker = blocker
// 		// daemon.EndBlocker = stop_blocker
// 	}
//
// 	logrus.Info("Register as service...")
// 	registerAsService(e)
//
// 	var http2 = true
// 	mode := vxconf.GetIntR("server.mode")
// 	switch mode {
// 	case 1:
// 		http2 = false
// 	case 2:
// 		http2 = true
// 	}
//
// 	if http2 {
// 		baseDir := path.Dir(vxconf.GetUsedConfigFile())
// 		fCert = strings.Replace(fCert, "{{.configDir}}", baseDir, -1)
// 		fKey = strings.Replace(fKey, "{{.configDir}}", baseDir, -1)
// 		// e.TLSServer.ErrorLog = log
// 		if err := s.cool.OnPreStart(e); err != nil {
// 			logrus.Errorf("ERROR: %v", err)
// 			logrus.Info("@@ exiting the server")
// 		} else if err := e.StartTLS(addr, fCert, fKey); err != nil {
// 			logrus.Errorf("ERROR: %v", err)
// 			logrus.Info("@@ shutting down the tls server")
// 		} else {
// 			fmt.Println("    stopped.")
// 		}
// 	} else {
// 		hs := &http.Server{
// 			Addr:         addr,
// 			ReadTimeout:  20 * time.Minute,
// 			WriteTimeout: 20 * time.Minute,
// 		}
// 		e.DisableHTTP2 = true
// 		if err := s.cool.OnPreStart(e); err != nil {
// 			logrus.Errorf("ERROR: %v", err)
// 			logrus.Info("@@ exiting the server")
// 		} else if err := e.StartServer(hs); err != nil {
// 			logrus.Errorf("ERROR: %v", err)
// 			logrus.Info("@@ shutting down the server")
// 		} else {
// 			fmt.Println("    stopped.")
// 		}
// 	}
// 	close(quit) // 正常结束
// }

func startEchoServer(e *echo.Echo, s *echoServerImpl, stopCh, doneCh chan struct{}) {
	go quitLoop(e, stopCh, doneCh)
	go runNew(s, e, stopCh, doneCh)
}

func quitLoop(e *echo.Echo, stopCh, doneCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			logrus.Debugf("...xs-server shutdown going on.")
			ctx, cancelFunc := context.WithTimeout(context.TODO(), 8*time.Second)
			defer cancelFunc()

			if err := e.Shutdown(ctx); err != nil {
				logrus.Fatal("Shutdown failed: ", err)
			} else {
				logrus.Info("Shutdown ok.")
			}

			<-doneCh
			return
		}
	}
}

func runNew(s *echoServerImpl, e *echo.Echo, stopCh, doneCh chan struct{}) {
	fAddr := vxconf.GetStringR("server.address", "")
	fPort := vxconf.GetIntR("server.port", 2300)
	fCert := vxconf.GetStringR("server.tls.cert", "")
	fKey := vxconf.GetStringR("server.tls.key", "")

	if s := os.Getenv("PORT"); len(s) > 0 {
		if port1, err := strconv.Atoi(s); err == nil {
			fPort = port1
		}
	}

	addr := fAddr
	if strings.Index(fAddr, ":") < 0 {
		addr = fmt.Sprintf("%s:%d", fAddr, fPort)
	}
	f := vxconf.GetBoolR("server.foreground", false)
	if f {
		s.cool.OnPrintBanner(addr)
	} else {
		// daemon.Target = e
		// daemon.Blocker = blocker
		// daemon.EndBlocker = stop_blocker
	}

	logrus.Info("Register as service...")
	registerAsService(e)

	var err error
	switch vxconf.GetIntR("server.mode", 2) {
	case 2:
		baseDir := path.Dir(cmdr.GetUsedConfigFile())
		fCert = os.ExpandEnv(strings.Replace(fCert, "{{.configDir}}", baseDir, -1))
		fKey = os.ExpandEnv(strings.Replace(fKey, "{{.configDir}}", baseDir, -1))
		if cmdr.FileExists(fCert) && cmdr.FileExists(fKey) {
			// e.TLSServer.ErrorLog = log
			if err = s.cool.OnPreStart(e); err != nil {
				logrus.Errorf("ERROR: %v", err)
				logrus.Info("@@ exiting the server")
			} else if err = e.StartTLS(addr, fCert, fKey); err != nil {
				logrus.Errorf("ERROR: %v", err)
				logrus.Info("@@ shutting down the tls server")
			} else {
				fmt.Println("    stopped.")
			}
			break
		}
		logrus.Warnf("RESTful: cert files not found, fallthrough to HTTP/1.1... (%v|%v)", fCert, fKey)
		fallthrough

	default:
		hs := &http.Server{
			Addr:         addr,
			ReadTimeout:  20 * time.Minute,
			WriteTimeout: 20 * time.Minute,
		}
		e.DisableHTTP2 = true
		if err = s.cool.OnPreStart(e); err != nil {
			logrus.Errorf("ERROR: %v", err)
			logrus.Info("@@ exiting the server")
		} else if err = e.StartServer(hs); err != nil {
			logrus.Errorf("ERROR: %v", err)
			logrus.Info("@@ shutting down the server")
		} else {
			fmt.Println("    stopped.")
		}
	}

	// close(quitCh)
}
