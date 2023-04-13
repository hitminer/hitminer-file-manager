package multibar

import "io"

type DefaultBar struct{}

func (b *DefaultBar) NewBarReader(reader io.Reader, size int64, description string) io.Reader {
	return reader
}

func (b *DefaultBar) Wait() {}
