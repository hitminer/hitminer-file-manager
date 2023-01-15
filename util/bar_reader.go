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
		mpb.WithRefreshRate(180 * time.Millisecond),
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
			decor.AverageETA(decor.ET_STYLE_GO),
			decor.Name(" ] "),
			decor.AverageSpeed(decor.UnitKiB, "% .2f"),
		))
	return bar.ProxyReader(reader)
}
