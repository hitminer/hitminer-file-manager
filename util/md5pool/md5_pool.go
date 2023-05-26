package md5pool

import (
	"crypto/md5"
	"hash"
	"sync"
)

var p = sync.Pool{
	New: func() interface{} {
		return md5.New()
	},
}

func New() hash.Hash {
	return p.Get().(hash.Hash)
}

func Put(h hash.Hash) {
	h.Reset()
	p.Put(h)
}
