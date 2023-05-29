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

func WriteErofs(path string) error {
	return os.WriteFile(filepath.Join(path, "mkfs.erofs.exe"), erofs, 0755)
}
