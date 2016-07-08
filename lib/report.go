package cmdlib

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	day time.Duration = time.Hour * 24
)

var (
	homeDir           = os.Getenv("HOME")
	initialReportLen  = 16864
	maximumLineLength = 256 * 1024
)

var magnitudes = []struct {
	mag  time.Duration
	name string
}{
	{day, "day"},
	{time.Hour, "hour"},
	{time.Minute, "minute"},
	{time.Second, "second"},
}

// FormatRelativeTime converts a duration to a string according to magnitudes above
func FormatRelativeTime(diff time.Duration) string {
	var ret string

	if diff.Seconds() < 1.0 {
		return "Just now"
	}

	for _, mag := range magnitudes {
		count := diff / mag.mag
		diff = diff % mag.mag
		if count > 0 {
			ret = ret + strconv.FormatInt(int64(count), 10) + " " +
				mag.name
			if count > 1 {
				ret = ret + "s"
			}
			ret = ret + " "
		}
	}

	return ret + "ago"
}

// FormatTime formats the given timestring (UNIX time) to human readable string
func FormatTime(timestr string) string {
	timeint, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		Warningln("could not parse %s to integer: %v", timestr, err)
		return timestr
	}
	tm := time.Unix(timeint, 0)
	diff := time.Since(tm)

	if diff.Hours() < 24.0*7 {
		return FormatRelativeTime(diff)
	}

	return tm.Format("2006-01-02T15:04:05")
}

// ParseCmdLogLine prepares a single line for display
func ParseCmdLogLine(line string, session string, regex *regexp.Regexp,
	out *[]string) {
	items := strings.SplitN(line, "\t", 3)

	// The format of the line is improper
	if len(items) != 3 {
		return
	}

	// If session filtering is used and session does not match
	if session != "" && session != items[1] {
		return
	}

	// If regex is given and it does not match
	if regex != nil && !regex.MatchString(items[2]) {
		return
	}

	items[0] = FormatTime(items[0])
	copy(*out, items)
}

// ParseArgs is extendable list of arguments for the parseCmdLog function
type ParseArgs struct {
	Session string
	Grep    string
	Pwd     bool
	Reverse bool
	Output  io.Writer
}

// ParseCmdLog Parses and prints out the command log from given
// reader. Possibly filter by session.
func ParseCmdLog(input io.Reader, arg ParseArgs) (err error) {
	reader := bufio.NewReaderSize(input, maximumLineLength)

	var re *regexp.Regexp
	re = nil
	if arg.Grep != "" {
		re, err = regexp.Compile(arg.Grep)
		if err != nil {
			return ErrorLn("Failed to compile regexp \"", arg.Grep, "\": ", err)
		}
	}

	// The format for the report structure:
	// for each element: timestring, session, command, [cwd]
	// If the strings in the element are empty, it has been filtered out
	report := make([][]string, initialReportLen)
	index := 0

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		}
		if err != nil {
			return ErrorLn("Error reading log: ", err)
		}
		if index >= len(report) {
			report = append(report, []string{})
		}
		report[index] = make([]string, 4)
		ParseCmdLogLine(string(line), arg.Session, re, &report[index])
		index = index + 1
	}

	if arg.Pwd {
		AddPwdsToReport(&report)
	}

	// Print out the report
	reportlen := len(report) - 1
	for idx := range report {
		pos := idx
		if arg.Reverse {
			pos = reportlen - idx
		}
		if report[pos] != nil && report[pos][0] != "" {
			line := ""
			if arg.Session == "" {
				line = report[pos][1] + " "
			}
			line = line + report[pos][0]
			if arg.Pwd {
				line = line + "\t" + report[pos][3]
			}
			line = line + "\t" + report[pos][2] + "\n"
			arg.Output.Write([]byte(line))
		}
	}
	return nil
}

// Heuristic to determine the current directory
func determineDirectory(previous string, cmd string) string {
	ret := ""

	if cmd == "s" {
		ret = ".."
	} else if strings.HasPrefix(cmd, "cd") {
		parts := strings.SplitN(cmd, " ", 2)
		if len(parts) == 1 {
			ret = homeDir
		} else {
			ret = parts[1]
		}
	} else if strings.HasPrefix(cmd, "Started shell session: ") {
		parts := strings.SplitN(cmd, " ", 4)
		ret = parts[3]
	}

	if filepath.IsAbs(ret) {
		return ret
	}

	return filepath.Clean(filepath.Join(previous, ret))
}

// AddPwdsToReport Add working directories to the report
func AddPwdsToReport(report *[][]string) {
	sessions := make(map[string][]*[]string)

	for idx, item := range *report {
		if item != nil && item[0] != "" {
			sessions[item[1]] = append(sessions[item[1]], &(*report)[idx])
		}
	}

	for _, items := range sessions {
		cwd := homeDir
		for _, item := range items {
			cwd = determineDirectory(cwd, (*item)[2])
			(*item)[3] = cwd
		}
	}
}
