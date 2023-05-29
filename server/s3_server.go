package server

import (
	"context"
	"time"
)

type Object struct {
	FullPath         string `json:"FullPath"`
	Name             string `json:"Name"`
	LastModified     string `json:"LastModified"`
	LastModifiedTime time.Time
	Size             int    `json:"Size"`
	IsDirectory      bool   `json:"IsDirectory"`
	ETag             string `json:"Etag"`
}

type S3Server interface {
	GetObjects(ctx context.Context, filePath, objectName string) error
	PutObjects(ctx context.Context, filePath, objectName string, erofs bool) error
	ListObjects(ctx context.Context, prefix, delimiter string) <-chan Object
	MakeDirectory(ctx context.Context, prefix string) error
	RemoveObjects(ctx context.Context, objectName string, recursive bool) error
	CopyObjects(ctx context.Context, from, to string, recursive bool) error
	GetError() error
}
