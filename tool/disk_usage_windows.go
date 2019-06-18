/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	disk.All = 0
	disk.Free = 0
	disk.Used = 0
	return
}
