package cmdlib

import (
	"bufio"
	"io"
)

var (
	maximumLineLength = 160 * 1024
)

type LineReader interface {
	ReadLine() (string, error)
}

type BufferedReader struct {
	reader *bufio.Reader
}

func NewBufferedReader(f io.Reader) (ret *BufferedReader) {
	return &BufferedReader{reader: bufio.NewReaderSize(f, maximumLineLength)}
}

func (f *BufferedReader) ReadLine() (string, error) {
	return f.reader.ReadString('\n')
}
