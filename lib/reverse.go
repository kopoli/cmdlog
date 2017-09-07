package cmdlib

import (
	"bytes"
	"fmt"
	"io"
)

type ReverseReader struct {
	fp  io.ReadSeeker
	buf []byte

	// Position of the buffer start in the file
	pos int64

	// Position in the buffer
	bufpos int

	len int64
}

func NewReverseReader(f io.ReadSeeker) (ret *ReverseReader, err error) {
	ret = &ReverseReader{
		fp:     f,
		buf:    make([]byte, maximumLineLength),
		pos:    -1,
		bufpos: 0,
	}

	ret.len, err = ret.fp.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	ret.pos = ret.len
	// ret.bufpos = len(ret.buf)-1

	return
}

func (r *ReverseReader) fillBuffer() (err error) {
	// No more data, return EOF
	if r.pos == 0 && r.bufpos == 0 {
		err = io.EOF
		return
	}

	// If all data has been read to the buffer
	if r.pos == 0 {
		return
	}

	readlen := len(r.buf) - r.bufpos

	fmt.Println("buflen", len(r.buf), "bufpos", r.bufpos, "readlen", readlen)

	// If less than buffer length of data is left in the file
	if r.pos < int64(readlen) {
		readlen = int(r.pos)
	}

	fmt.Println("Readlen", readlen)

	// If would need to read more bytes than there is space in the
	// buffer
	if readlen == 0 {
		panic(fmt.Sprint("A line detected that is longer than ",
			maximumLineLength, "bytes"))
	}

	// If there is data still left in the buffer
	if r.bufpos > 0 {
		copy(r.buf[readlen:], r.buf[:r.bufpos-1])
	}

	// Set the read position to the new place
	r.pos, err = r.fp.Seek(int64(-readlen), io.SeekCurrent)
	if err != nil {
		return
	}

	// Read data to the front part of the buffer
	var n int
	n, err = r.fp.Read(r.buf[:readlen])
	if err != nil {
		return
	}
	if n != readlen {
		// TODO proper error
		panic(fmt.Sprint("There should have been", readlen,
			"bytes of data, but only", n, "was read"))
	}
	r.bufpos = readlen + r.bufpos
	// r.pos = r.pos - int64(readlen)
	return
}

func (r *ReverseReader) ReadLine() (line string, err error) {

	idx := bytes.LastIndex(r.buf[:r.bufpos], []byte{'\n'})

	fmt.Println("Found newline at", idx, "buffer len", r.bufpos, "input pos",
		r.pos)
	// if not found in the current buffer
	if idx < 0 {
		err = r.fillBuffer()
		if err != nil {
			return
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
	if idx > 0 {
		idx += 1
	}

	fmt.Printf("luetaan slice: %d:%d\n", idx, r.bufpos)
	line = string(r.buf[idx:r.bufpos])
	r.bufpos = idx
	if idx > 0 {
		r.bufpos -= 1
	}

	return
}

// func (r *Reverser) Read(p []byte) (n int, err error) {
// 	readlen := int64(len(p))
// 	tgt := &p
// 	smaller := false

// 	if r.pos-readlen < 0 {
// 		readlen = r.pos
// 		tmp := make([]byte, readlen)
// 		tgt = &tmp
// 		smaller = true
// 	}

// 	if readlen == 0 {
// 		return
// 	}

// 	r.pos, err = r.fp.Seek(-readlen, io.SeekCurrent)
// 	if err != nil {
// 		// TODO debug
// 		panic("Reverse reader failed!")
// 	}

// 	n, err = r.fp.Read(*tgt)
// 	if err != nil {
// 		return
// 	}

// 	if n != int(readlen) {
// 		panic(fmt.Sprint("Readlen differs:", readlen, "!=", n))
// 	}

// 	if smaller {
// 		copy(p, *tgt)
// 	}
// 	return
// }

// func (r *Reverser) Close() error {
// 	return r.fp.Close()
// }
