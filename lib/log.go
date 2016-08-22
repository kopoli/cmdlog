package cmdlib

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	filters = []string{
		" *ls? -[thlroa]* *$",
		" *l[shla]* *$",
	}
)

// AppendLine creates a log line to the given logfile
func AppendLine(logfile io.Writer, session string, args []string) {
	cmd := strings.Join(args, " ")

	// change to single line command
	re := regexp.MustCompile("[\r\n]+")
	cmd = re.ReplaceAllString(cmd, " ")

	// Filter out unlogged commands
	for _, filter := range filters {
		re := regexp.MustCompile(filter)
		if re.MatchString(cmd) {
			return
		}
	}

	fmt.Fprintf(logfile, "%d\t%s\t%s\n", time.Now().Unix(), session, cmd)
	return
}
