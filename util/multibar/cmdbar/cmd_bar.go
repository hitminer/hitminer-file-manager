package cmdbar

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
	"time"
)

type CmdBar struct {
	p *mpb.Progress
}

func NewBar(w io.Writer) *CmdBar {
	return &CmdBar{
		p: mpb.New(
			mpb.WithRefreshRate(200*time.Millisecond),
			mpb.WithOutput(w),
		),
	}
}

func (b *CmdBar) NewBarReader(reader io.Reader, size int64, description string) io.Reader {
	bar := b.p.New(size,
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

func (b *CmdBar) Wait() {
	time.Sleep(300 * time.Millisecond)
	//b.p.Shutdown()
}
