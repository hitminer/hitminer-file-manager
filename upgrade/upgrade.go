package upgrade

import (
	"context"
	"fmt"
	"github.com/hitminer/hitminer-file-manager/util/multibar/cmdbar"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Upgrade(ctx context.Context, w io.Writer) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	name := filepath.Base(executable)
	tempPath := filepath.Join(filepath.Dir(executable), fmt.Sprintf(".%s.new", name))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	f, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tempPath)
	}()

	b := cmdbar.NewBar(w)
	bar := b.NewBarReader(resp.Body, resp.ContentLength, "upgrade")
	_, err = io.Copy(f, bar)
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, executable)
	if err != nil {
		return err
	}
	return nil
}
