package cmdlib

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkBufferedReader(b *testing.B) {
	var r *BufferedReader
	var err error

	maximumLineLength = 1024

	r = NewBufferedReader(bytes.NewReader([]byte(testData)))

	for i := 0; i < b.N; i++ {
		for {
			_, err = r.ReadLine()
			if err == io.EOF {
				break
			}
		}
	}
}
