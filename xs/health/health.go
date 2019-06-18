/*
 * Copyright © 2019 Hedzr Yeh.
 */

package health

import (
	"github.com/hedzr/cmdr/conf"
	voxr_common "github.com/hedzr/voxr-common"
	"github.com/labstack/echo"
	"time"
)

type (
	Handlers struct {
		// DB *mgo.Session
		// DB    *dbi.Config
		// mgoDB *mgo.Session
	}

	Health struct {
		Status  bool    `json:"status"`
		Msg     string  `json:"msg"`
		App     *string `json:"app"`
		Ver     *string `json:"ver"`
		Githash *string `json:"githash"`
		Builtts *string `json:"builtts"`
	}
)

func Enable(e *echo.Echo) {
	e.Any("/health", healthFunc)
	e.Any(voxr_common.GetApiPrefix()+"/health", healthFunc)
}

func Update(c func(*Health)) {
	c(health)
}

// SetUpdater checks in an updater with health-check period.
// It does just work for consul health checking mechanism.
func SetUpdater(updater func(h *Health, tm *time.Time)) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			// TODO retrieve the global (from daemon) stopCh and wait for it.
			// case _ = <-voxr_common.GetAppExitCh():
			// 	return
			case tm := <-ticker.C:
				updater(health, &tm)
			}
		}
	}()
}

func healthFunc(c echo.Context) (err error) {
	err = c.JSON(200, &health)
	return
}

//
var health = &Health{true, "OK",
	&conf.AppName, &conf.Version,
	&conf.Githash, &conf.Buildstamp}
