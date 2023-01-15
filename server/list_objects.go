package server

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go/v7"
	"strings"
	"time"
)

func (svr *Server) ListObjects(ctx context.Context, prefix string) string {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range svr.client.ListObjects(ctx, svr.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: false}) {
			if object.Err != nil {
				svr.errChan <- object.Err
			} else {
				objectsCh <- object
			}
		}
	}()

	loc := time.Now().Location()
	var builder strings.Builder
	for obj := range objectsCh {
		if prefix == obj.Key {
		} else if strings.HasSuffix(obj.Key, "/") {
			builder.WriteString("drwxr-xr-x\t")
			builder.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(obj.Size))))
			builder.WriteString(obj.LastModified.In(loc).Format("Jan _2 15:04"))
			builder.WriteString("\t")
			builder.WriteString(obj.Key[len(prefix) : len(obj.Key)-1])
			builder.WriteString("\n")
		} else {
			builder.WriteString("-rwxr-xr--\t")
			builder.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(obj.Size))))
			builder.WriteString(obj.LastModified.In(loc).Format("Jan 02 15:04"))
			builder.WriteString("\t")
			builder.WriteString(obj.Key[len(prefix):])
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
