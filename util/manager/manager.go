package manager

import (
	"context"
	"errors"
	"sync"
)

type Manager struct {
	limit   chan struct{}
	wg      *sync.WaitGroup
	errWg   *sync.WaitGroup
	errChan chan error
	err     error
}

func NewManager(ctx context.Context, limit int) *Manager {
	mg := &Manager{
		limit:   make(chan struct{}, limit),
		wg:      &sync.WaitGroup{},
		errWg:   &sync.WaitGroup{},
		errChan: make(chan error, 1),
		err:     nil,
	}
	mg.errWg.Add(1)
	go func() {
		defer mg.errWg.Done()
		for {
			select {
			case v, ok := <-mg.errChan:
				if !ok {
					return
				}
				if v != nil {
					mg.err = errors.Join(mg.err, v)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return mg
}

func (mg *Manager) AppendError(err error) {
	mg.errChan <- err
}

func (mg *Manager) GetError() error {
	return mg.err
}

func (mg *Manager) Add() {
	mg.limit <- struct{}{}
	mg.wg.Add(1)
}

func (mg *Manager) Done() {
	<-mg.limit
	mg.wg.Done()
}

func (mg *Manager) Wait() {
	mg.wg.Wait()
}

func (mg *Manager) Finish() {
	mg.Wait()
	close(mg.errChan)
	mg.errWg.Wait()
}
