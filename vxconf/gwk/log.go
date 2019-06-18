/*
 * Copyright © 2019 Hedzr Yeh.
 */

package gwk

import (
	"github.com/sirupsen/logrus"
)

var debug bool

func info(msg string) {
	if debug {
		logrus.Print(msg)
	}
}

func infof(fmt string, args ...interface{}) {
	if debug {
		logrus.Printf(fmt, args...)
	}
}

func warn(msg string) {
	if debug {
		logrus.Print(msg)
	}
}

func warnf(fmt string, args ...interface{}) {
	if debug {
		logrus.Printf(fmt, args...)
	}
}

func Error(msg string) {
	if debug {
		logrus.Print(msg)
	}
}

func Errorf(fmt string, args ...interface{}) {
	if debug {
		logrus.Printf(fmt, args...)
	}
}

func fatal(msg string) {
	logrus.Fatal(msg)
}

func fatalf(fmt string, args ...interface{}) {
	logrus.Fatalf(fmt, args...)
}
