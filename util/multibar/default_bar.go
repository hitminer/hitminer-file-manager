package multibar

import "io"

type DefaultBar struct{}

func (b *DefaultBar) Write(p []byte) (n int, err error) { return 0, nil }

func (b *DefaultBar) NewCntBar(size int64, description string) {}

func (b *DefaultBar) SetPrint(print bool) {}

func (b *DefaultBar) NewBarReader(reader io.Reader, size int64, description string) io.Reader {
	return reader
}

func (b *DefaultBar) Wait() {}
