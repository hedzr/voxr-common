/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}
