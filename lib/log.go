package cmdlib

import (
	"fmt"
	"io"
	"regexp"
	"time"
)

var (
	filters = []string{
		" *ls? -[thlroa]* *$",
		" *l[shla]* *$",
	}
)

// AppendLine creates a log line to the given logfile
func AppendLine(logfile io.Writer, session string, args string) error {

	// change to single line command
	re := regexp.MustCompile("[\r\n]+")
	args = re.ReplaceAllString(args, " ")

	// Filter out unlogged commands
	for _, filter := range filters {
		re := regexp.MustCompile(filter)
		if re.MatchString(args) {
			return nil
		}
	}

	_, err := fmt.Fprintf(logfile, "%d\t%s\t%s\n", time.Now().Unix(), session, args)
	return err
}
