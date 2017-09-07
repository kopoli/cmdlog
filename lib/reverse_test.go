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
	tests := []struct {
		name    string
		data    string
		lines   []string
		wantErr bool
	}{
		{"Empty data", "", []string{}, false},
		{"Single row", "a", []string{"a"}, false},
		{"Two rows", "a\nb", []string{"b","a"}, false},
		{"Empty line at end", "a\n", []string{"","a"}, false},
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
			for {
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

			if len(tt.lines) != len(lines) {
				t.Errorf("Expected %d lines, got %d lines", len(tt.lines), len(lines))
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
