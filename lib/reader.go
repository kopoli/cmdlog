package cmdlib

import (
	"bufio"
	"io"
)

type LineReader interface {
	ReadLine() (string, error)
}

type BufferedReader struct {
	reader *bufio.Reader
}

func NewBufferedReader(f io.Reader, maximumLineLength int) (ret *BufferedReader) {
	return &BufferedReader{reader: bufio.NewReaderSize(f, maximumLineLength)}
}

func (f *BufferedReader) ReadLine() (string, error) {
	return f.reader.ReadString('\n')
}
