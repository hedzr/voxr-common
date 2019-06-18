/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

import (
	"math/rand"
	"strconv"
)

// Intv 忽略任何错误转换字符串为整数值并返回。如果无法转换，返回值为0
func Intv(s string) (v int) {
	v, _ = strconv.Atoi(s)
	return
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandRandSeq() string {
	n := rand.Intn(128)
	return RandSeq(n)
}
