//go:build darwin && arm64
// +build darwin,arm64

package ero

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed static/darwin_arm64/mkfs.erofs
var erofs []byte

func WriteErofs(path string) error {
	return os.WriteFile(filepath.Join(path, "mkfs.erofs"), erofs, 0755)
}
