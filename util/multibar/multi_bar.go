package multibar

import "io"

type MultiBar interface {
	io.Writer
	NewCntBar(size int64, description string)
	SetPrint(print bool)
	NewBarReader(reader io.Reader, size int64, description string) io.Reader
	Wait()
}
