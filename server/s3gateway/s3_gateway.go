package s3gateway

import (
	"context"
	"github.com/hitminer/hitminer-file-manager/util/manager"
	"github.com/hitminer/hitminer-file-manager/util/multibar"
)

type S3Server struct {
	host  string
	token string
	mg    *manager.Manager
	bar   multibar.MultiBar
}

func NewS3Server(ctx context.Context, host, token string, bar multibar.MultiBar) *S3Server {
	if bar == nil {
		bar = &multibar.DefaultBar{}
	}
	return &S3Server{
		host:  host,
		token: token,
		mg:    manager.NewManager(ctx, 8),
		bar:   bar,
	}
}
