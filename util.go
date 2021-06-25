package main

import "os"

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
