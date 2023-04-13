package multibar

import "io"

type MultiBar interface {
	NewBarReader(reader io.Reader, size int64, description string) io.Reader
	Wait()
}
