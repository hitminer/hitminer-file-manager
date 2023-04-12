package multibar

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
	"time"
)

var p *mpb.Progress
var closed bool

func init() {
	p = mpb.New(
		mpb.WithRefreshRate(200 * time.Millisecond),
	)
	closed = false
}

func NewBarReader(reader io.Reader, size int64, description string) io.Reader {
	if closed {
		return reader
	}
	bar := p.New(size,
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Name(description, decor.WCSyncSpaceR),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		))
	return bar.ProxyReader(reader)
}

func Close() {
	closed = true
	p = nil
}

func Wait() {
	time.Sleep(300 * time.Millisecond)
	// p.Wait()
}
