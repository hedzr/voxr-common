/*
 * Copyright © 2019 Hedzr Yeh.
 */

package vxconf

import (
	"context"
	"github.com/hedzr/voxr-common/db/dbi"
	"github.com/hedzr/voxr-common/xs/mjwt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// var (
// 	CfgFile string
// 	AppName string
//
// 	Silent     bool = false // suppress the logger information (all log info will be discarded rather than output to stderr)
// 	Verbose    bool = false // verbose logging
// 	VerboseVV  bool = false
// 	VerboseVVV bool = false
// 	Debug      bool = false // debug mode
//
// 	// these 3 variables will be rewrote when app had been building by ci-tool
// 	Version    = "0.1.7"
// 	Buildstamp = ""
// 	Githash    = ""
//
// 	ServerTag = ""
// 	ServerID  = ""
//
// 	RealStart                 func()                 = nil
// 	Deregister                func()                 = nil
// 	PrintVersion              func()                 = nil
// 	OnConfigChanged           func(e fsnotify.Event) = nil
// 	OnConfigReloadedFunctions                        = make(map[OnConfigReloaded]bool)
// )

type (
	// OnConfigReloaded interface {
	// 	OnConfigReloaded()
	// }

	CoolServer interface {
		GetApiPrefix() string
		GetApiVersion() string

		OnPrintBanner(addr string)

		OnInitRegistrar()
		OnShutdownRegistrar()

		OnInitCORS(*middleware.CORSConfig) echo.MiddlewareFunc

		OnInitJWT(e *echo.Echo, config mjwt.JWTConfig) mjwt.JWTConfig
		JWTSkipper(c echo.Context) bool

		OnInitForwarders(e *echo.Echo)
		OnShutdownForwarders()

		OnInitStats(e *echo.Echo) (ready bool)
		OnInitStatic(e *echo.Echo) (ready bool)
		OnInitTemplates(e *echo.Echo) (ready bool)

		OnInitDB(e *echo.Echo) *dbi.Config
		OnShutdownDB()

		OnInitRoutes(e *echo.Echo) (ready bool)
		OnInitWebSocket(e *echo.Echo) (ready bool)

		OnPreStart(e *echo.Echo) (err error)
		OnPreShutdown()
		OnPostShutdown()
	}

	BlockingFunc func()

	XsServer interface {
		Start(stopCh, doneCh chan struct{}) BlockingFunc
		Stop(cxt context.Context) error
		Restart() BlockingFunc

		// Deregister() error

		// OnInitCORS()
		// OnInitRoutes()
	}
)
