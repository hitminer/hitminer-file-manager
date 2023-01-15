package server

import (
	"context"
	"github.com/minio/minio-go/v7"
	"hitminer-file-manager/util"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

func (svr *Server) FPutObjects(ctx context.Context, filePath, objectName string) {
	stat, err := os.Stat(filePath)
	if err != nil {
		svr.errChan <- err
		return
	}
	if !stat.IsDir() {
		svr.Add()
		go func() {
			defer svr.Done()
			if strings.HasSuffix(objectName, "/") {
				objectName += filepath.Base(filePath)
			}
			err := svr.fPutObjects(ctx, filePath, filepath.ToSlash(objectName))
			if err != nil {
				svr.errChan <- err
			}
		}()
		return
	}
	infos, err := os.ReadDir(filePath)
	if err != nil {
		svr.errChan <- err
		return
	}
	svr.Add()
	go func() {
		defer svr.Done()
		dirName := objectName
		if !strings.HasSuffix(dirName, "/") {
			dirName += "/"
		}
		_, err := svr.client.PutObject(ctx, svr.bucket, dirName, nil, 0, minio.PutObjectOptions{})
		if err != nil {
			svr.errChan <- err
		}
	}()
	for _, info := range infos {
		if info.IsDir() {
			svr.FPutObjects(ctx, filepath.Join(filePath, info.Name()), filepath.ToSlash(filepath.Join(objectName, info.Name())))
		} else {
			svr.Add()
			filePath, objectName := filepath.Join(filePath, info.Name()), filepath.ToSlash(filepath.Join(objectName, info.Name()))
			go func() {
				defer svr.Done()
				err := svr.fPutObjects(ctx, filePath, objectName)
				if err != nil {
					svr.errChan <- err
				}
			}()
		}
	}
}

func (svr *Server) fPutObjects(ctx context.Context, filePath, objectName string) error {
	fileReader, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = fileReader.Close()
	}()

	fileStat, err := fileReader.Stat()
	if err != nil {
		return err
	}

	fileSize := fileStat.Size()
	bar := util.NewBarReader(fileReader, fileSize, "upload: "+filePath)
	opts := minio.PutObjectOptions{}
	if opts.ContentType = mime.TypeByExtension(filepath.Ext(filePath)); opts.ContentType == "" {
		opts.ContentType = "application/octet-stream"
	}
	_, err = svr.client.PutObject(ctx, svr.bucket, objectName, bar, fileSize, opts)
	return err
}
