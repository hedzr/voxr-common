/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"context"
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/hedzr/voxr-common/xs/skippers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

type (
	echoServerImpl struct {
		e    *echo.Echo
		cool vxconf.CoolServer
		// stopCh chan struct{}
		// doneCh chan struct{}
	}
)

// var stopChX = make(chan struct{}, 5)

func (s *echoServerImpl) Start(stopCh, doneCh chan struct{}) vxconf.BlockingFunc {
	e := s.initServer()

	// s.initJWT(e)
	// s.initGZIP(e)
	s.initContextReplacer(e)
	s.initStats(e)
	s.cool.OnInitForwarders(e)
	// s.initStatic(e)
	s.initTemplates(e)

	s.initLogger(e)
	s.initBodyLimit(e)
	s.initCORS(e)
	s.initCSRF(e)
	s.initGzip(e)
	s.initJWT(e)
	s.initSecure(e)

	logrus.Debugf("in echoServerImpl Start()")
	// e.Logger.Debugf("command-line options: %v\n", conf.AppConfigSource.Results)

	db := s.cool.OnInitDB(e)

	s.initStatic(e)
	s.initRoutes(e, db)
	s.initWebSocket(e)

	startEchoServer(e, s, stopCh, doneCh)

	return s.blocker(stopCh)
}

// func (s *echoServerImpl) stopBlocker() {
// 	stopChX <- struct{}{}
// 	time.Sleep(1 * time.Second)
// }

func (s *echoServerImpl) blocker(stopCh chan struct{}) func() {
	return func() {
		<-stopCh

		logrus.Info("@@ shutting down")

		_ = s.Stop(context.TODO()) // stop daemon

		s.doShutdown(s.e)

		logrus.Info("@@ All folks.")
	}
}

func (s *echoServerImpl) initServer() *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	e.HTTPErrorHandler = s.MyDefaultHTTPErrorHandler

	e.Pre(middleware.MethodOverride())

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// initLogger(e)
	// initBodyLimit(e)
	// initCORS(s, e)
	// initCSRF(e)
	// initGzip(e)
	// initJWT(e)
	//
	// initSecure(e)
	// initStatic(e)
	// initRoutes(e)
	// initWebSocket(e)

	// beforeRun(e)

	// run(e, common.AppExitCh)

	// port := vxconf.GetIntR("server.port")
	// e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))

	logrus.Info("Using logger level ", vxconf.GetStringR("server.logger.level", "OFF"))
	return e
}

func (s *echoServerImpl) initBodyDump(e *echo.Echo) {
	if vxconf.GetBoolR("server.bodyDump.enabled", false) {
		// e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		// 	Handler: func(c echo.Context, reqBody []byte, resBody []byte) {
		// 		e.Logger.Debugf("-body-dump-: req:\n%v\nres:\n%v", reqBody, resBody)
		// 	},
		// }))
		e.Use(middleware.BodyDump(func(c echo.Context, reqBody []byte, resBody []byte) {
			e.Logger.Debugf("-body-dump-: req:\n%v\nres:\n%v", reqBody, resBody)
		}))
	}
}

func (s *echoServerImpl) initLogger(e *echo.Echo) {
	// e.Use(middleware.Logger())

	e.Logger = NewEchoLogger()
	// e.Logger.SetLevel(log.ERROR)
	// e.Logger.SetLevel(log.INFO)
	e.Logger.SetLevel(stringToLevel(vxconf.GetStringR("server.logger.level", "OFF")))
	e.Logger.Debugf("initLogger() for echo.")

	// e.Use(middleware.Logger())
}

func (s *echoServerImpl) initBodyLimit(e *echo.Echo) {

	// fBodyLimit := vxconf.GetStringR("server.bodyLimit")
	// if len(fBodyLimit) == 0 {
	// 	fBodyLimit = "5M"
	// }
	// e.Use(middleware.BodyLimit(fBodyLimit))

	e.Use(middleware.BodyLimit(vxconf.GetStringR("server.bodyLimit", "4M")))
	// e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{}))
}

func (s *echoServerImpl) initCORS(e *echo.Echo) {
	origins := vxconf.GetStringSliceR("server.cors.origins", nil)
	selfUrl := fmt.Sprintf("%s://%s:%v", "http", "localhost", vxconf.GetIntR("server.port", 2300))
	found := false
	for _, url := range origins {
		if url == selfUrl {
			found = true
			break
		}
	}
	if !found {
		origins = append(origins, selfUrl)
	}
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: origins,
	// 	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	// }))

	e.Use(s.cool.OnInitCORS(&middleware.CORSConfig{
		AllowOrigins: origins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	})) // https://echo.labstack.com/middleware/cors
	logrus.Infof("CORS: %v", vxconf.GetStringSliceR("server.cors.origins", nil))

}

func (s *echoServerImpl) initCSRF(e *echo.Echo) {
	if vxconf.GetBoolR("server.csrf.enabled", false) {
		tokenLookup := vxconf.GetStringR("server.csrf.tokenLookup", "header:X-XSRF-TOKEN")
		// e.Use(middleware.CSRF()) // https://echo.labstack.com/middleware/csrf
		e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLookup: tokenLookup,
		}))
	}
}

func (s *echoServerImpl) initGzip(e *echo.Echo) {
	if vxconf.GetBoolR("server.gzip.enabled", false) {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level:   vxconf.GetIntR("server.gzip.level", 5),
			Skipper: skippers.DefaultGzipSkipper,
		}))
	}
}

func (s *echoServerImpl) initSecure(e *echo.Echo) {
	if vxconf.GetBoolR("server.secure.enabled", true) {
		e.Use(middleware.Secure()) // https://echo.labstack.com/middleware/secure
	}
}

func ds(val, defval string) string {
	if len(val) == 0 {
		return defval
	}
	return val
}

func (s *echoServerImpl) initStatic(e *echo.Echo) {
	if s.cool.OnInitStatic(e) == true {
		return
	}

	// default impl

	runmode := vxconf.GetStringR("runmode", "devel")
	root := vxconf.GetStringR("server.static.root", "/var/lib/$APPNAME/public") // webui.StaticPagesRootDir()
	index := vxconf.GetStringR("server.static.index", "index.html")             // webui.StaticPagesIndexFile()
	urlPrefix := vxconf.GetStringR("server.static.urlPrefix", "/public")        // webui.StaticPagesUrlPrefix()
	list := vxconf.GetBoolR("server.static.list", true)                         // webui.StaticPagesUrlPrefix()
	forceList := vxconf.GetBoolR("server.static.forceList", false)              // webui.StaticPagesUrlPrefix()

	if runmode == "prod" {
		if !forceList {
			list = false
		}
	}

	root = os.ExpandEnv(root)
	appName := vxconf.GetStringR("server.serviceName", conf.AppName)

	if strings.Contains(root, "{{.AppName}}") {
		root = strings.Replace(root, "{{.AppName}}", appName, -1)
	}

	if !cmdr.FileExists(root) {
		savedRoot := root
		root = path.Join(tool.GetCurrentDir(), appName, urlPrefix)
		if !cmdr.FileExists(root) {
			root = path.Join(tool.GetCurrentDir(), strings.TrimLeft(appName, "vx-"), urlPrefix)
			if !cmdr.FileExists(root) {
				root = path.Join(tool.GetCurrentDir(), urlPrefix) // 用和 urlPrefix 相同的名字
			}
		}
		logrus.Infof("static folder `%s` not exists: using local root = %s | serviceName = %s", savedRoot, root, appName)
	}

	e.Group(urlPrefix, middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   root,
		Index:  index,
		Browse: list,
		HTML5:  vxconf.GetBoolR("server.static.html5", true), // true if u want SPA like,
	}))
	// Another
	// e.Static("/", "public")
	// Another:
	// assetHandler := http.FileServer(rice.MustFindBox("app").HTTPBox())
	// e.GET("/", echo.WrapHandler(assetHandler))
	// e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))

	logrus.Infof("`static` ready: url'%s' => '%s'", urlPrefix, root)
}

func (s *echoServerImpl) initRoutes(e *echo.Echo, dbConfig *dbi.Config) {
	if s.cool.OnInitRoutes(e) == false {
		s.initDefaultRoutes(e, dbConfig)
	}
	logrus.Info("Routes ready...")
}

func (s *echoServerImpl) initWebSocket(e *echo.Echo) (ready bool) {
	if s.cool.OnInitWebSocket(e) == false {
		// ... default websocket actions here
	}
	ready = true
	return
}

//

//

func (s *echoServerImpl) initContextReplacer(e *echo.Echo) {
	// e.Use(gwk.ContextReplacer)
}

func (s *echoServerImpl) initStats(e *echo.Echo) {
	// if s.cool.OnInitStats(e) == true {
	// 	return
	// }
	//
	// // default impl
	//
	// ss := stats.NewStats()
	// e.Use(ss.Process)
	// e.GET(s.cool.GetApiPrefix()+"/stats", ss.Handle)   // Endpoint to get stats
	// e.GET(s.cool.GetApiPrefix()+"/metrics", ss.Handle) // Endpoint to get stats
	// e.Use(stats.ServerHeader)
}

func (s *echoServerImpl) initTemplates(e *echo.Echo) {
	if s.cool.OnInitTemplates(e) == true {
		return
	}
}

func (s *echoServerImpl) initDefaultRoutes(e *echo.Echo, db *dbi.Config) {

	// // JSONP
	// e.GET("/pri/jsonp", func(c echo.Context) error {
	// 	callback := c.QueryParam("callback")
	// 	var content struct {
	// 		Response  string    `json:"response"`
	// 		Timestamp time.Time `json:"timestamp"`
	// 		Random    int       `json:"random"`
	// 	}
	// 	content.Response = "Sent via JSONP"
	// 	content.Timestamp = time.Now().UTC()
	// 	content.Random = rand.Intn(1000)
	// 	return c.JSONP(http.StatusOK, callback, &content)
	// })
	//
	// e.GET(s.cool.GetApiPrefix()+"/foo", func(c echo.Context) error {
	// 	cc := c.(*gwk.Context)
	// 	return c.String(http.StatusOK, fmt.Sprintf("Hello, World! [%s]", cc.Foo()))
	// })
	//
	// // https://github.com/labstack/echox/blob/master/cookbook/streaming-response/server.go
	// e.GET("/pri/stream", stream.Handle)
	//
	// // WebSocket handler
	// // https://github.com/labstack/echox/blob/master/cookbook/reverse-proxy/upstream/server.go
	// e.GET("/pri/ws", func(c echo.Context) error {
	// 	name := "any-web-socket"
	// 	if len(c.QueryParams().Get("name")) > 0 {
	// 		name = c.QueryParams().Get("name")
	// 	}
	// 	websocket.Handler(func(ws *websocket.Conn) {
	// 		defer ws.Close()
	// 		for {
	// 			// Write
	// 			err := websocket.Message.Send(ws, fmt.Sprintf("Hello from upstream server %s!", name))
	// 			if err != nil {
	// 				e.Logger.Error(err)
	// 			}
	// 			time.Sleep(1 * time.Second)
	// 		}
	// 	}).ServeHTTP(c.Response(), c.Request())
	// 	return nil
	// })

	logrus.Info("Default Routes ready: /foo, /pri/stream, /pri/ws, /pri/jsonp ...")
	// https://github.com/labstack/echox/blob/master/cookbook/file-upload/multiple/server.go
	// https://github.com/labstack/echox/blob/master/cookbook/file-upload/single/server.go
}

func (s *echoServerImpl) doShutdown(e *echo.Echo) {
	//
	// config, server.forwarders Shutdown,
	// s.cool.AppExitCh <- true
	// close(s.cool.AppExitCh)

	// cancelHttpServer(e)

	s.cool.OnPreShutdown()

	logrus.Infof("@@ shutting down registrar.")
	s.cool.OnShutdownRegistrar()
	Deregister()

	logrus.Infof("@@ shutting down forwarders.")
	s.cool.OnShutdownForwarders()

	logrus.Infof("@@ shutting down db.")
	s.cool.OnShutdownDB()

	s.cool.OnPostShutdown()
	// <-lastQuit
}

// func (s *echoServerImpl) startAndBlockOld(e *echo.Echo) {
// 	// Start server
// 	// e.Logger.Fatal(e.Start(":1323"))
// 	// Start server
// 	//e.Logger.Info("starting the server at :1323")
// 	quit := quitSignal()
// 	go run(nil, e, quit)
//
// 	if vxconf.GetBoolR("server.foreground") {
// 		fmt.Println("Press CONTROL-C to exit.")
// 	}
// 	<-quit
//
// 	logrus.Infof("@@ os interrupt signal catched, server will be shutdowning gracefully.")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer func() {
// 		fmt.Println("    10 seconds waited.")
// 		cancel()
// 		fmt.Println("    cancel() called.")
// 	}()
// 	if err := e.Shutdown(ctx); err != nil {
// 		e.Logger.Fatal(err)
// 	} else {
//
// 		// config, server.forwarders Shutdown,
// 		s.doShutdown(e)
//
// 		fmt.Println("    shutdown() called.")
// 	}
// 	logrus.Info("@@ All folks.")
// }

func (s *echoServerImpl) Stop(cxt context.Context) error {
	if vxconf.GetBoolR("server.foreground", false) == false {
		// daemon.Stop()
	}
	return nil
}

// func (s *echoServerImpl) StopOld() error {
// 	pidPath := tools.DefaultPidPath(vxconf.GetStringR("server.serviceName"))
// 	if pid, err := strconv.Atoi(tools.ReadContent(pidPath)); err != nil {
// 		logrus.Fatalf("ERROR: %v", err)
// 	} else if pid > 0 {
// 		signal := syscall.SIGTSTP
// 		if vxconf.GetBoolR("server.stop.hup") {
// 			signal = syscall.SIGHUP
// 		}
// 		if vxconf.GetBoolR("server.stop.kill") {
// 			signal = syscall.SIGKILL
// 		}
// 		if err = syscall.Kill(pid, signal); err != nil {
// 			logrus.Fatalf("ERROR: %v", err)
// 			return err
// 		}
// 		fmt.Printf("service %d running, signal triggered.\n", pid)
// 	}
// 	return nil
// }

func (s *echoServerImpl) Restart() vxconf.BlockingFunc {
	if vxconf.GetBoolR("server.foreground", false) == false {
		// daemon.Restart()
	}
	return func() {}
}

// func (s *echoServerImpl) RestartOld() (error, cli_common.BlockingFunc) {
// 	if err := s.Stop(); err != nil {
// 		logrus.Fatalf("FATAL: %v", err)
// 		return nil, nil
// 	} else {
// 		return nil, s.Start()
// 	}
// }
