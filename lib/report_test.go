package cmdlib

import (
	"regexp"
	"testing"
)

func BenchmarkParseCmdLogLine(b *testing.B) {
	out := make([]string, 4)
	for i := 0; i < b.N; i++ {
		ParseCmdLogLine("1450120005	zsh-2755-20151214	go test",
			"", nil, &out)
	}
}

func BenchmarkParseCmdLogLine_RegexpMatch(b *testing.B) {
	out := make([]string, 4)
	re := regexp.MustCompile("go test")
	for i := 0; i < b.N; i++ {
		ParseCmdLogLine("1450120005	zsh-2755-20151214	go test",
			"", re, &out)
	}
}
