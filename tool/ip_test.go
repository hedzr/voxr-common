/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool_test

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestLookupHostInfo(t *testing.T) {
	host, _ := os.Hostname()
	{
		addrs, _ := net.LookupIP(host)
		for _, addr := range addrs {
			if ipv4 := addr.To4(); ipv4 != nil {
				fmt.Println("IPv4: ", ipv4)
			}
		}
	}

	fmt.Println("test more...")

	{
		addrs, _ := net.InterfaceAddrs()
		fmt.Printf("%v\n", addrs)
		for _, addr := range addrs {
			fmt.Printf("IPv4: %v | %v / %v\n", addr, addr.String(), addr.Network())
		}
	}
}
