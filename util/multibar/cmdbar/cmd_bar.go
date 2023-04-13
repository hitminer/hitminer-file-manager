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
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		))
	return bar.ProxyReader(reader)
}

func (b *CmdBar) Wait() {
	time.Sleep(300 * time.Millisecond)
	// p.Wait()
}
