package server

import (
	"context"
	"github.com/minio/minio-go/v7"
	"hitminer-file-manager/util"
	"os"
)

func (svr *Server) SolveConflict(ctx context.Context, filePath, objectName string) {
	dirs := util.SplitFullPath(objectName)
	for i, dir := range dirs {
		_, err := svr.client.StatObject(ctx, svr.bucket, dir, minio.StatObjectOptions{})
		if err == nil {
			err := svr.client.RemoveObject(ctx, svr.bucket, dir, minio.RemoveObjectOptions{GovernanceBypass: true})
			if err != nil {
				svr.errChan <- err
			}
		}

		if i != len(dirs)-1 {
			_, err := svr.client.StatObject(ctx, svr.bucket, dir+"/", minio.StatObjectOptions{})
			if err != nil {
				_, err := svr.client.PutObject(ctx, svr.bucket, dir+"/", nil, 0, minio.PutObjectOptions{})
				if err != nil {
					svr.errChan <- err
					return
				}
			}
		}
	}

	// 删除本身
	err := svr.client.RemoveObject(ctx, svr.bucket, objectName, minio.RemoveObjectOptions{GovernanceBypass: true})
	if err != nil {
		svr.errChan <- err
		return
	}

	if filePath != "" {
		stat, err := os.Stat(filePath)
		if err != nil {
			svr.errChan <- err
			return
		}
		if !stat.IsDir() {
			svr.RemoveObjects(ctx, objectName)
		}
	}
}
