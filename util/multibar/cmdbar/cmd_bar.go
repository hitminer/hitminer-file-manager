package cmdbar

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
	"time"
)

type CmdBar struct {
	p      *mpb.Progress
	cntBar *mpb.Bar
	print  bool
}

func NewBar(w io.Writer) *CmdBar {
	return &CmdBar{
		p: mpb.New(
			mpb.WithRefreshRate(200*time.Millisecond),
			mpb.WithOutput(w),
		),
		cntBar: nil,
		print:  true,
	}
}

func (b *CmdBar) Write(p []byte) (n int, err error) {
	if b.cntBar != nil {
		b.cntBar.Increment()
	}
	return 0, nil
}

func (b *CmdBar) NewCntBar(size int64, description string) {
	bar := b.p.New(size,
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Name(description, decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
		))
	b.cntBar = bar
}

func (b *CmdBar) SetPrint(print bool) {
	b.print = print
}

func (b *CmdBar) NewBarReader(reader io.Reader, size int64, description string) io.Reader {
	if !b.print {
		return reader
	}
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
