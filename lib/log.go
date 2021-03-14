package cmdlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	defaultFilters = []string{
		"^ *ls? -[thlroa]* *$",
		"^ *l[shla]* *$",
	}
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return ! (err != nil && os.IsNotExist(err))
}


type Log struct {
	LogFile string
	Filters []string
	FilterFile string
}

func CreateLog(logfile, filterfile string) (*Log) {
	ret := &Log{
		LogFile: logfile,
		Filters: defaultFilters,
		FilterFile: filterfile,
	}

	return ret
}

// AppendLine creates a log line to the given logfile
func (l *Log)AppendLine(session string, args string) error {

	// change to single line command
	re := regexp.MustCompile("[\r\n]+")
	args = re.ReplaceAllString(args, " ")

	// delete trailing whitespace
	args = strings.TrimRight(args, " ")

	// Filter out unlogged commands
	for _, filter := range l.Filters {
		re := regexp.MustCompile(filter)
		if re.MatchString(args) {
			return nil
		}
	}

	fp, err := os.OpenFile(l.LogFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = fmt.Fprintf(fp, "%d\t%s\t%s\n", time.Now().Unix(),
		session, args)
	if err != nil {
		return err
	}

	return fp.Close()
}

// SaveDefaultFilters saves the default filters as an example if such file
// does not yet exist.
func (l *Log)SaveDefaultFilters() error {
	if FileExists(l.FilterFile) {
		return nil
	}

	sb := strings.Builder{}

	_, _ = sb.WriteString(`# cmdlog log line filter file. One regular expression filter per line.
# Syntax: empty, whitespace and lines starting with # are ignored.
`)
	for _, filter := range defaultFilters {
		_, _ = sb.WriteString(filter)
		_, _ = sb.WriteString("\n")
	}

	out := []byte(sb.String())
	return ioutil.WriteFile(l.FilterFile, out, 0600)
}

func (l *Log)LoadFilters() error {
	var err error

	// Do nothing if filter file is not found
	if !FileExists(l.FilterFile) {
		return nil
	}

	emptyLineRe := regexp.MustCompile(`^\s*(#.*)?$`)

	fp, err := os.Open(l.FilterFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	filterErrors := strings.Builder{}

	reader := NewBufferedReader(fp)

	linenum := 0
	filters := []string{}
	done := false
	for !done {
		linenum++
		line, err := reader.ReadLine()

		if err == io.EOF {
			done = true
			err = nil
		}
		if err != nil {
			return err
		}
		if emptyLineRe.MatchString(line) {
			continue
		}

		_, err = regexp.Compile(line)
		if err != nil {
			filterErrors.WriteString(fmt.Sprintf(
				"%s:%d: Invalid regexp: \"%s\": %v\n",
				l.FilterFile, linenum, line, err))
		} else {
			filters = append(filters, strings.TrimSuffix(line, "\n"))
		}
	}

	l.Filters = filters

	if len(filterErrors.String()) > 0 {
		return fmt.Errorf("Parsing filters failed: %s", filterErrors.String())
	}

	return nil
}
