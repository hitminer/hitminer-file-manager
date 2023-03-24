package util

import (
	"os"
	"path/filepath"
	"sync"
)

func RandFile(fileName string, size int) {
	f, _ := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	defer func() {
		_ = f.Close()
	}()
	_, _ = f.WriteString(RandString(size))
}

func RandDir(dir string, size int, fileNames ...string) {
	_ = os.MkdirAll(dir, 0755)
	wg := sync.WaitGroup{}

	for _, fileName := range fileNames {
		fileName := fileName
		wg.Add(1)
		go func() {
			RandFile(filepath.Join(dir, fileName), size)
			wg.Done()
		}()
	}
	wg.Wait()
}

func DeleteDir(dir string) {
	_ = os.RemoveAll(dir)
}
