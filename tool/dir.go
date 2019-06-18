/*
 * Copyright © 2019 Hedzr Yeh.
 */

package tool

import (
	"log"
	"os"
	"path/filepath"
)

func GetExcutableDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(dir)
	return dir
}

func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(dir)
	return dir
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
