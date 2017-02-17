package cmdlib

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"testing"
)

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
		Grep: "dpkg",
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

