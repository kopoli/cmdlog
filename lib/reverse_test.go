package cmdlib

import (
	"bytes"
	"io"
	"testing"
)

func TestReverseReader(t *testing.T) {
	type fields struct {
		fp     io.ReadSeeker
		buf    []byte
		pos    int64
		bufpos int
		len    int64
	}

	maximumLineLength = 10

	tests := []struct {
		name    string
		data    string
		lines   []string
		wantErr bool
	}{
		{"Empty data", "", []string{}, false},
		{"Single row", "a", []string{"a\n"}, false},
		{"Two rows", "a\nb", []string{"b\n", "a\n"}, false},
		{"Empty line at end", "a\n", []string{"\n", "a\n"}, false},
		{"Longer lines", "something\nother\nthan\nthis",
			[]string{"this\n", "than\n", "other\n", "something\n"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *ReverseReader
			var err error
			r, err = NewReverseReader(bytes.NewReader([]byte(tt.data)))
			if err != nil {
				t.Errorf("Creating new reader failed:", err)
				return
			}
			lines := []string{}
			for i := 0; i < len(tt.lines)+1; i++ {
				line, err := r.ReadLine()
				if err == io.EOF {
					break
				}
				if (err != nil) != tt.wantErr {
					t.Errorf("ReverseReader.ReadLine() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				lines = append(lines, line)
			}

			_, err = r.ReadLine()
			if err != io.EOF {
				t.Errorf("Expected EOF when ReadLine called %d'th time",
					len(tt.lines)+2)
			}

			if len(tt.lines) != len(lines) {
				t.Errorf("Expected %d lines, got %d lines, Contents:\n%s",
					len(tt.lines), len(lines), diffStr(tt.lines,lines))
				return

			}
			for i := range tt.lines {
				if !structEquals(tt.lines[i], lines[i]) {
					t.Errorf("Expected line: %v\nGot line: %v",
						tt.lines[i], lines[i])
				}
			}
		})
	}
}

func BenchmarkReverseReader(b *testing.B) {
	var r *ReverseReader
	var err error

	maximumLineLength = 1024

	r, err = NewReverseReader(bytes.NewReader([]byte(testData)))
	if err != nil {
		b.Errorf("Creating a reverse reader from testData failed:",err)
		return
	}

	for i := 0; i < b.N; i++ {
		for {
			_, err = r.ReadLine()
			if err == io.EOF {
				break
			}
		}
	}
}
