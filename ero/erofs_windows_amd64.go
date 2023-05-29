//go:build windows && amd64
// +build windows,amd64

package ero

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed static/windows_amd64/mkfs.erofs.exe
var erofs []byte

//go:embed static/windows_amd64/cygwin1.dll
var cygwin1 []byte

func WriteErofs(path string) error {
	if err := os.WriteFile(filepath.Join(path, "mkfs.erofs.exe"), erofs, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "cygwin1.dll"))
}
