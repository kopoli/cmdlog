package cmdlib

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestAppendLine(t *testing.T) {
	var buf bytes.Buffer

	session := "ses"
	cmd := "first second"

	AppendLine(&buf, session, cmd)
	items := strings.Split(buf.String(), "\t")
	if len(items) != 3 {
		t.Error("Erroneous number of arguments (should be", 3,
			"is", len(items), ")")
	}

	_, err := strconv.ParseInt(items[0], 10, 64)
	if err != nil {
		t.Error("Could not parse", items[0], "as integer:", err)
	}

	if items[1] != session {
		t.Error("Improper session", items[1], "!=", session)
	}

	if strings.TrimRight(items[2], "\n") != cmd {
		t.Error("Command is improperly concatenated:", items[2],
			"Opposed to", cmd)
	}
}

func TestAppendLongLine(t *testing.T) {
	var buf bytes.Buffer

	var cmd string
	count := 32 * 1024
	str := "KEKE"
	for i := 0; i < count; i++ {
		cmd = cmd + " " + str
	}
	AppendLine(&buf, "Jeps", cmd)
	items := strings.SplitN(buf.String(), "\t", 3)
	length := count*len(str) + count + 1
	if len(items[2]) != length {
		t.Error("Length of the command line is", len(items[2]),
			"should be", length)
	}
}
