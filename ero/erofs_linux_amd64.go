//go:build linux && amd64
// +build linux,amd64

package ero

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed static/linux_amd64/mkfs.erofs
var erofs []byte

func WriteErofs(path string) error {
	return os.WriteFile(filepath.Join(path, "mkfs.erofs"), erofs, 0755)
}
