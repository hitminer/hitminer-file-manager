package server

import (
	"context"
	"github.com/minio/minio-go/v7"
	"strings"
)

func (svr *Server) MakeDirectory(ctx context.Context, prefix string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	_, err := svr.client.PutObject(ctx, svr.bucket, prefix, nil, 0, minio.PutObjectOptions{})
	if err != nil {
		svr.errChan <- err
	}
}
