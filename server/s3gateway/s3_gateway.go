package s3gateway

import (
	"context"
	"hitminer-file-manager/util/manager"
)

type S3Server struct {
	host  string
	token string
	mg    *manager.Manager
}

func NewS3Server(ctx context.Context, host, token string) *S3Server {
	return &S3Server{
		host:  host,
		token: token,
		mg:    manager.NewManager(ctx, 8),
	}
}
