package cmdlib

import (
	"bufio"
	"os"
)

const (
	maximumLineLength = 256 * 1024
)

type LineReader interface {
	ReadLine() (string, error)
}

type FileReader struct {
	reader *bufio.Reader
}

func NewFileReader(f *os.File) (ret *FileReader) {
	return &FileReader{reader: bufio.NewReaderSize(f, maximumLineLength)}
}

func (f *FileReader) ReadLine() (string, error) {
	return f.reader.ReadString('\n')
}
