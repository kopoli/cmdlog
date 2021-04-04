package cmdlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

type testLineReader struct {
	buf *bytes.Buffer
}

func (l *testLineReader) ReadLine() (string, error) {
	return l.buf.ReadString('\n')
}

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

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		diff time.Duration
		out  string
	}{
		{0, "Just now"},
		{time.Second, "1s ago"},
		{time.Second * 2, "2s ago"},
		{time.Second * 60, "1m ago"},
		{time.Hour * 2, "2h ago"},
		{time.Hour*2 + time.Second, "2h 1s ago"},
		{time.Hour*2 + time.Second + time.Minute*3, "2h 3m 1s ago"},
		{time.Hour * 24, "1d ago"},
		{time.Hour * 24 * 7, "7d ago"},
		{time.Hour * 24 * 31, "31d ago"},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%d -> %s", tt.diff, tt.out)
		t.Run(name, func(t *testing.T) {
			out := FormatRelativeTime(tt.diff)
			compare(t, "Relative time differs", tt.out, out)
		})
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
		{"Empty", "", ParseArgs{}, "", false},
		{"Invalid regexp", "", ParseArgs{Grep: "["}, "", true},
		{"Invalid time", "", ParseArgs{Since: "jeejee"}, "", true},
		{"Single line", "0\tsession\tcmdline\n",
			ParseArgs{Control: controlArgs{
				Now:      time.Unix(0, 0),
			}},
			"session Just now\tcmdline\n", false},

		{"Multiple lines", `0	session	cmdline
1450120005	zsh-2755-20151214	go test
`, ParseArgs{}, `session 1970-01-01T02:00:00	cmdline
zsh-2755-20151214 2015-12-14T21:06:45	go test
`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &testLineReader{buf: bytes.NewBufferString(tt.input)}
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
	input := &testLineReader{buf: bytes.NewBufferString(testData)}
	pa := ParseArgs{
		Output: ioutil.Discard,
		Grep:   "dpkg",
	}

	for i := 0; i < b.N; i++ {
		_ = ParseCmdLog(input, pa)
	}
}

func BenchmarkParseCmdLog_Whole(b *testing.B) {
	input := &testLineReader{buf: bytes.NewBufferString(testData)}
	pa := ParseArgs{
		Output: ioutil.Discard,
	}

	for i := 0; i < b.N; i++ {
		_ = ParseCmdLog(input, pa)
	}
}

func BenchmarkParseCmdLogLineNoAlloc(b *testing.B) {
	out := make([]string, 4)
	now := time.Time{}
	for i := 0; i < b.N; i++ {
		ParseCmdLogLineNoAlloc("1450120005	zsh-2755-20151214	go test",
			"", 0, now, nil, &out)
	}
}

func BenchmarkParseCmdLogLineNoAlloc_RegexpMatch(b *testing.B) {
	out := make([]string, 4)
	re := regexp.MustCompile("go test")
	now := time.Time{}
	for i := 0; i < b.N; i++ {
		ParseCmdLogLineNoAlloc("1450120005	zsh-2755-20151214	go test",
			"", 0, now, re, &out)
	}
}

func BenchmarkFormatRelativeTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatRelativeTime(time.Second * time.Duration(i))
	}
}
func TestParseCmdLogLineNoAlloc(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		session string
		since   int64
		now     time.Time
		regex   *regexp.Regexp
		out     []string
	}{
		{"Normal line", "1450120005	zsh-2755-20151214	go test", "", 0, time.Now(), nil,
			[]string{"2015-12-14T21:06:45", "zsh-2755-20151214", "go test"}},
	}
	out := []string{"", "", ""}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ParseCmdLogLineNoAlloc(tt.line, tt.session, tt.since, tt.now, tt.regex, &out)
		})
		for i := range out {
			if tt.out[i] != out[i] {
				t.Error("Invalid field", i, "Expected:", tt.out[i], "Got:", out[i])
			}
		}
	}
}

func Test_determineDirectory(t *testing.T) {
	tests := []struct {
		name     string
		previous string
		cmd      string
		want     string
	}{
		{"Home", "/something", "cd", homeDir},
		{"Go to subdir", "/something", "cd jeejee", "/something/jeejee"},
		{"Go to absdir", "/something", "cd /abs", "/abs"},
		{"Go up", "/something", "s", "/"},
		{"Go up 2", "/something", "cd ..", "/"},
		{"Start session", "/", "Started shell session: /here", "/here"},
		{"Normal command", "/something", "ls", "/something"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineDirectory(tt.previous, tt.cmd); got != tt.want {
				t.Errorf("determineDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
