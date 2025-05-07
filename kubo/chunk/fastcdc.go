package chunk

import (
	"io"

	"github.com/restic/chunker"
)

// FastCDC implements the Splitter interface using FastCDC.
type FastCDC struct {
	ch     *chunker.Chunker
	reader io.Reader
}

// NewFastCDC creates a new FastCDC chunker using restic/chunker.
func NewFastCDC(r io.Reader, avgBlkSize uint64) Splitter {
	const (
		MinSize = 512
		MaxSize = 64 * 1024
	)
	const poly = chunker.Pol(0x3DA3358B4DC173)

	// Note: restic/chunker only uses min and max
	ch := chunker.NewWithBoundaries(r, poly, MinSize, MaxSize)

	return &FastCDC{
		ch:     ch,
		reader: r,
	}
}

func (f *FastCDC) NextBytes() ([]byte, error) {
	chunk, err := f.ch.Next(nil)
	if err != nil {
		return nil, err
	}
	return chunk.Data, nil
}

func (f *FastCDC) Reader() io.Reader {
	return f.reader
}