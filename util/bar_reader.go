package util

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
	"time"
)

var p *mpb.Progress

func init() {
	p = mpb.New(
		mpb.WithRefreshRate(200 * time.Millisecond),
	)
}

func NewBarReader(reader io.Reader, size int64, description string) io.Reader {
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
