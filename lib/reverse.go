package cmdlib

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ReverseReader struct {
	fp  io.ReadSeeker
	buf []byte

	// Position of the buffer start in the file
	pos int64

	// Position in the buffer
	bufpos int
}

func NewReverseReader(f io.ReadSeeker) (ret *ReverseReader, err error) {
	ret = &ReverseReader{
		fp:     f,
		buf:    make([]byte, maximumLineLength),
		pos:    -1,
		bufpos: 0,
	}

	ret.pos, err = ret.fp.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	return
}

func (r *ReverseReader) fillBuffer() error {
	var err error
	// No more data, return EOF
	if r.pos == 0 && r.bufpos == 0 {
		err = io.EOF
		return err
	}

	// If all data has been read to the buffer
	if r.pos == 0 {
		return nil
	}

	readlen := len(r.buf) - r.bufpos

	// If less than buffer length of data is left in the file
	if r.pos < int64(readlen) {
		readlen = int(r.pos)
	}

	// If would need to read more bytes than there is space in the
	// buffer
	if readlen == 0 {
		panic(fmt.Sprint("A line detected that is longer than ",
			maximumLineLength, "bytes"))
	}

	// If there is data still left in the buffer
	if r.bufpos > 0 {
		copy(r.buf[readlen:], r.buf[:r.bufpos])
	}

	// Set the read position to the new place
	r.pos, err = r.fp.Seek(r.pos-int64(readlen), io.SeekStart)
	if err != nil {
		return err
	}

	// Read data to the front part of the buffer
	var n int
	n, err = r.fp.Read(r.buf[:readlen])
	if err != nil {
		return err
	}
	if n != readlen {
		// TODO proper error
		panic(fmt.Sprint("There should have been", readlen,
			"bytes of data, but only", n, "was read"))
	}
	r.bufpos = readlen + r.bufpos
	return nil
}

func (r *ReverseReader) ReadLine() (line string, err error) {
	idx := bytes.LastIndex(r.buf[:r.bufpos], []byte{'\n'})

	// if not found in the current buffer
	if idx < 0 {
		err = r.fillBuffer()
		if err != nil {
			return "", err
		}
		idx = bytes.LastIndex(r.buf[:r.bufpos], []byte{'\n'})
		if idx == -1 {
			if r.pos > 0 {
				panic(fmt.Sprint("2ND A line detected that is longer than ",
					maximumLineLength, "bytes"))
			}
			idx = 0
		}
	}
	if idx > 0 || (idx == 0 && r.buf[idx] == '\n') {
		idx++
	}

	sb := strings.Builder{}
	sb.Write(r.buf[idx:r.bufpos])
	sb.WriteByte('\n')

	r.bufpos = idx
	if idx > 0 || (idx == 0 && r.buf[idx] == '\n') {
		r.bufpos--
	}

	return sb.String(), nil
}
