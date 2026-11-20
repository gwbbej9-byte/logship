//go:build !windows

package main

import "os"

func openFollow(path string) (*os.File, error) {
	return os.Open(path)
}
