package cmdlib

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestBufferedReader(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		length  int
		wantErr bool
	}{
		// data without newline produces io.EOF error and the data
		{"Empty data", "", 0, true},
		{"Single row without newline", "abc", 10, true},
		{"Single row without newline", "abc", 1, true},
		{"Single row with newline", "abc\n", 10, false},
		// The underlying buffer expands automatically
		{"Too short line for the buffer",
			"123456789012345678\n", 16, false},
	}

	var r *BufferedReader

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r = NewBufferedReader(bytes.NewReader([]byte(tt.data)), tt.length)

			s, err := r.ReadLine()
			if tt.wantErr != (err != nil) {
				t.Fatalf("Wanted error: %v got error: %v", tt.wantErr, err)
			}
			ret := strings.SplitN(tt.data, "\n", 2)
			s = strings.TrimSuffix(s, "\n")
			if s != ret[0] {
				t.Fatalf("Expected string: [%v]\nGot string: [%v]",
					ret[0], s)
			}
		})
	}
}

func BenchmarkBufferedReader(b *testing.B) {
	var r *BufferedReader
	var err error

	maximumLineLength := 1024

	r = NewBufferedReader(bytes.NewReader([]byte(testData)), maximumLineLength)

	for i := 0; i < b.N; i++ {
		for {
			_, err = r.ReadLine()
			if err == io.EOF {
				break
			}
		}
	}
}
