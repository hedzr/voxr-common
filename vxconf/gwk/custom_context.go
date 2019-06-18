/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"github.com/labstack/echo"
)

type (
	Context struct {
		echo.Context
	}
)

func ContextReplacer(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &Context{c}
		return h(cc)
	}
}

func (c *Context) Foo() string {
	println("foo")
	return "foo"
}

func (c *Context) Bar() string {
	println("bar")
	return "bar"
}
