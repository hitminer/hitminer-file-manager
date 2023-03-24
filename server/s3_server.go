package server

import (
	"context"
	"io"
)

type S3Server interface {
	GetObjects(ctx context.Context, filePath, objectName string) error
	PutObjects(ctx context.Context, filePath, objectName string) error
	ListObjects(ctx context.Context, prefix string, out io.Writer) error
	MakeDirectory(ctx context.Context, prefix string) error
	RemoveObjects(ctx context.Context, objectName string, recursive bool) error
	CopyObjects(ctx context.Context, from, to string, recursive bool) error
}
