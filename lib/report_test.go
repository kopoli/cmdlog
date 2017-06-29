package cmdlib

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

func structEquals(a, b interface{}) bool {
	return spew.Sdump(a) == spew.Sdump(b)
}

func diffStr(a, b interface{}) (ret string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(spew.Sdump(a)),
		B:        difflib.SplitLines(spew.Sdump(b)),
		FromFile: "Expected",
		ToFile:   "Received",
		Context:  3,
	}

	ret, _ = difflib.GetUnifiedDiffString(diff)
	return
}

func compare(t *testing.T, msg string, a, b interface{}) {
	if !structEquals(a, b) {
		t.Error(msg, "\n", diffStr(a, b))
	}
}

func TestParseCmdLog(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		arg     ParseArgs
		output  string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := bytes.NewBufferString(tt.input)
			buf := &bytes.Buffer{}
			tt.arg.Output = buf
			if err := ParseCmdLog(input, tt.arg); (err != nil) != tt.wantErr {
				t.Fatalf("ParseCmdLog() error = %v, wantErr %v", err, tt.wantErr)
			}

			compare(t, "Outputs differ", tt.output, buf.String())
		})
	}
}

var testData string = `
1463382163	zsh-5604-20160516	Started shell session
1463382167	zsh-5604-20160516	Exited shell session
1463382174	zsh-16168-20160503	gobu -help
1463382195	zsh-26914-20160504	gobu debug linux nocgo shrink trimpath race
1463382200	zsh-26914-20160504	ls -lh cmdlog
1463382253	zsh-26914-20160504	./cmdlog -version
1463382279	zsh-26914-20160504	./cmdlog -help
1463382323	zsh-26914-20160504	./cmdlog report -help
1463382327	zsh-26914-20160504	./cmdlog report -reverse -grep ls
1463382333	zsh-26914-20160504	thelm --title cmdlog --hide-initial --single-arg ./cmdlog report --reverse --grep
`

func BenchmarkParseCmdLog_Regexp(b *testing.B) {
	buf := bytes.NewBufferString(testData)
	pa := ParseArgs{
		Output: ioutil.Discard,
		Grep:   "dpkg",
	}

	for i := 0; i < b.N; i++ {
		ParseCmdLog(buf, pa)
	}
}

func BenchmarkParseCmdLog_Whole(b *testing.B) {
	buf := bytes.NewBufferString(testData)
	pa := ParseArgs{
		Output: ioutil.Discard,
	}

	for i := 0; i < b.N; i++ {
		ParseCmdLog(buf, pa)
	}
}

func BenchmarkParseCmdLogLine(b *testing.B) {
	out := make([]string, 4)
	for i := 0; i < b.N; i++ {
		ParseCmdLogLine("1450120005	zsh-2755-20151214	go test",
			"", 0, nil, &out)
	}
}

func BenchmarkParseCmdLogLine_RegexpMatch(b *testing.B) {
	out := make([]string, 4)
	re := regexp.MustCompile("go test")
	for i := 0; i < b.N; i++ {
		ParseCmdLogLine("1450120005	zsh-2755-20151214	go test",
			"", 0, re, &out)
	}
}
