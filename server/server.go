package server

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/multierr"
	"sync"
)

type Server struct {
	belong  string
	client  *minio.Client
	bucket  string
	limit   chan int
	wg      *sync.WaitGroup
	errWg   *sync.WaitGroup
	errChan chan error
	Err     error
}

func NewServer(ctx context.Context, endpoint, accessKey, secretKey, belong string) *Server {
	client, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
		Region: "us-east-1",
	})
	svr := &Server{
		belong:  belong,
		client:  client,
		bucket:  "workplace",
		limit:   make(chan int, 8),
		wg:      &sync.WaitGroup{},
		errWg:   &sync.WaitGroup{},
		errChan: make(chan error),
		Err:     nil,
	}
	svr.errWg.Add(1)
	go func() {
		defer svr.errWg.Done()
		for {
			select {
			case v, ok := <-svr.errChan:
				if !ok {
					return
				}
				if v != nil {
					svr.Err = multierr.Append(svr.Err, v)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return svr
}

func (svr *Server) Add() {
	svr.limit <- 0
	svr.wg.Add(1)
}

func (svr *Server) Done() {
	<-svr.limit
	svr.wg.Done()
}

func (svr *Server) Wait() {
	svr.wg.Wait()
}

func (svr *Server) Finish() {
	close(svr.errChan)
	svr.errWg.Wait()
}
