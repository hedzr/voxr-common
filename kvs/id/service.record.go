/*
 * Copyright © 2019 Hedzr Yeh.
 */

package id

import (
	"fmt"
	"net"
)

func GenerateServerRecordId(ip net.IP, port int) string {
	return fmt.Sprintf("%s:%d", ip.String(), port)
}
