package server

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"strings"
)

func (svr *Server) RemoveObjects(ctx context.Context, prefix string) {
	if strings.HasSuffix(prefix, "/") {
		err := svr.client.RemoveObject(ctx, svr.bucket, prefix[:len(prefix)-1], minio.RemoveObjectOptions{GovernanceBypass: true})
		if err != nil {
			svr.errChan <- err
		}
	}

	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range svr.client.ListObjects(ctx, svr.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				svr.errChan <- object.Err
			} else {
				objectsCh <- object
			}
		}
	}()

	for rErr := range svr.client.RemoveObjects(ctx, svr.bucket, objectsCh, minio.RemoveObjectsOptions{GovernanceBypass: true}) {
		if rErr.Err != nil {
			svr.errChan <- rErr.Err
		}
	}
}

func (svr *Server) RemoveObject(ctx context.Context, prefix string) {
	dir := prefix
	if !strings.HasSuffix(prefix, "/") {
		dir = prefix + "/"
	}

	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range svr.client.ListObjects(ctx, svr.bucket, minio.ListObjectsOptions{Prefix: dir, Recursive: true, MaxKeys: 5}) {
			if object.Err != nil {
				svr.errChan <- object.Err
			} else {
				objectsCh <- object
			}
		}
	}()

	var err error
	for range objectsCh {
		err = fmt.Errorf("cannot remove '%s': Is a directory", prefix[len(svr.belong):])
	}
	if err != nil {
		svr.errChan <- err
		return
	}

	err = svr.client.RemoveObject(ctx, svr.bucket, prefix, minio.RemoveObjectOptions{GovernanceBypass: true})
	if err != nil {
		svr.errChan <- err
	}
}
