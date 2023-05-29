//go:build darwin && amd64
// +build darwin,amd64

package ero

import "fmt"

func WriteErofs(path string) error {
	return fmt.Errorf("unsupport os and arch")
}
