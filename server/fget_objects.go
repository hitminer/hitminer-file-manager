package server

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"hitminer-file-manager/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (svr *Server) FGetObjects(ctx context.Context, filePath, objectName string) {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range svr.client.ListObjects(ctx, svr.bucket, minio.ListObjectsOptions{Prefix: objectName, Recursive: true}) {
			if object.Err != nil {
				svr.errChan <- object.Err
			} else {
				objectsCh <- object
			}
		}
	}()

	if strings.HasSuffix(filePath, separator) {
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			svr.errChan <- err
			return
		}
	}

	for obj := range objectsCh {
		objT := obj
		if strings.HasSuffix(objT.Key, "/") {
			// 不是本身
			if objT.Key != objectName+"/" {
				err := os.MkdirAll(filepath.Join(filePath, objT.Key[len(objectName):len(objT.Key)-1]), 0755)
				if err != nil {
					svr.errChan <- err
				}
			} else {
				err := os.MkdirAll(filePath, 0755)
				if err != nil {
					svr.errChan <- err
				}
			}
			continue
		}
		svr.Add()
		go func() {
			defer svr.Done()
			localPath := ""
			if objT.Key == objectName {
				if strings.HasSuffix(filePath, separator) {
					localPath = filePath + filepath.Base(objectName)
				} else {
					localPath = filePath
				}
			} else {
				localPath = filepath.Join(filePath, objT.Key[len(objectName):])
			}
			st, err := os.Stat(localPath)
			if err == nil {
				if st.IsDir() {
					svr.errChan <- fmt.Errorf("fileName: %s is a directory", localPath)
					return
				}
			}
			ret, err := svr.client.GetObject(ctx, svr.bucket, objT.Key, minio.GetObjectOptions{})
			if err != nil {
				svr.errChan <- err
				return
			}
			f, err := os.OpenFile(localPath, os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				svr.errChan <- err
				return
			}
			defer func() {
				_ = f.Close()
			}()
			stat, err := ret.Stat()
			if err != nil {
				svr.errChan <- err
				return
			}
			bar := util.NewBarReader(ret, stat.Size, "download: "+localPath)
			_, err = io.Copy(f, bar)
			if err != nil {
				svr.errChan <- err
				return
			}

		}()
	}
}
