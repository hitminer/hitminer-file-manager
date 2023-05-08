package upgrade

import (
	"context"
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
	stat, err := os.Stat(executable)
	if err != nil {
		return err
	}

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

	tempDir, err := os.MkdirTemp("", "hitminer")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	downloadFile, err := os.OpenFile(filepath.Join(tempDir, filepath.Base(downloadUrl)), os.O_RDWR|os.O_CREATE|os.O_EXCL, stat.Mode())
	if err != nil {
		return err
	}

	b := cmdbar.NewBar(w)
	bar := b.NewBarReader(resp.Body, resp.ContentLength, "upgrade")
	_, err = io.Copy(downloadFile, bar)
	if err != nil {
		_ = downloadFile.Close()
		return err
	}

	err = downloadFile.Close()
	if err != nil {
		return err
	}

	oldName := filepath.Join(tempDir, name+".old")
	err = os.MkdirAll(filepath.Dir(oldName), 0755)
	if err != nil {
		return err
	}
	err = os.Rename(executable, oldName)
	if err != nil {
		return err
	}

	err = os.Rename(downloadFile.Name(), executable)
	if err != nil {
		return err
	}
	return nil
}
