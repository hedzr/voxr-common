/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

import (
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	"net/http"
)

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (s *echoServerImpl) MyDefaultHTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	logrus.Debugf("err, %s", err)
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else if c.Echo().Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			if err := c.NoContent(code); err != nil {
				goto ERROR
			}
		} else {
			if err := c.JSON(code, msg); err != nil {
				goto ERROR
			}
		}
	}
ERROR:
	logrus.Errorf("[MyDefaultHTTPErrorHandler] err: %v; req: %v", err, c.Path())
}
