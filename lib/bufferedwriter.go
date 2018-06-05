package cmdlib

import (
	"bufio"
	"io"
)

type BufferedWriter struct {
	buf     *bufio.Writer
	counter int
	lines   int
}

func NewBufferedWriter(o io.Writer, lines int) *BufferedWriter {
	return &BufferedWriter{
		buf: bufio.NewWriter(o),
	}
}

func (b *BufferedWriter) Write(p []byte) (n int, err error) {
	n, err = b.buf.Write(p)

	b.counter++
	if b.counter == b.lines {
		err = b.buf.Flush()
		b.counter = 0
	}
	return n, err
}

func (b *BufferedWriter) Close() error {
	return b.buf.Flush()
}
