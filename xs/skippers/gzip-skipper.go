/*
 * Copyright © 2019 Hedzr Yeh.
 */

package skippers

import "github.com/labstack/echo"

// DefaultGzipSkipper returns false which processes the middleware.
func DefaultGzipSkipper(c echo.Context) bool {
	return false
}
