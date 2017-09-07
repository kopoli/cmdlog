package cmdlib

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	day time.Duration = time.Hour * 24
)

var (
	homeDir          = os.Getenv("HOME")
	initialReportLen = 1024 * 128
	timeFormat       = "2006-01-02T15:04:05"
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
func FormatTime(timeint int64) string {
	tm := time.Unix(timeint, 0)
	diff := time.Since(tm)

	if diff.Hours() < 24.0*7 {
		return FormatRelativeTime(diff)
	}

	return tm.Format(timeFormat)
}

// ParseCmdLogLine prepares a single line for display
func ParseCmdLogLine(line string, session string, since int64, regex *regexp.Regexp,
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

	timeint, err := strconv.ParseInt(items[0], 10, 64)
	if err != nil {
		items[0] = "<invalid>"
	} else if timeint < since {
		return
	} else {
		items[0] = FormatTime(timeint)
	}

	copy(*out, items)
}

// ParseCmdLogLineNoAlloc prepares a single line without unnecessary allocation.
func ParseCmdLogLineNoAlloc(line string, session string, since int64, regex *regexp.Regexp,
	out *[]string) {

	var pos [2]int
	start := 0
	for i := range pos {
		// if start >= len(line) {
		// 	panic(fmt.Sprint("report line is of invalid format: length: ",
		// 		len(line), " Contents:", line))
		// 	// return
		// }
		relpos := strings.Index(line[start:], "\t")

		// The format of the line is improper
		if relpos < 0 {
			return
		}

		pos[i] = relpos + start + 1
		start = pos[i] + 1
	}

	// If session filtering is used and session does not match
	if session != "" && session != line[pos[0]:pos[1]-1] {
		return
	}

	// If regex is given and it does not match
	if regex != nil && !regex.MatchString(line[pos[1]:]) {
		return
	}

	timeint, err := strconv.ParseInt(line[:pos[0]-1], 10, 64)
	if err != nil {
		(*out)[0] = "<invalid>"
	} else if timeint < since {
		return
	} else {
		(*out)[0] = FormatTime(timeint)
	}

	(*out)[1] = line[pos[0] : pos[1]-1]
	(*out)[2] = line[pos[1]:]
}

// ParseArgs is extendable list of arguments for the parseCmdLog function
type ParseArgs struct {
	Session string
	Since   string
	Grep    string
	Pwd     bool
	Reverse bool
	Output  io.Writer
}

// ParseCmdLog Parses and prints out the command log from given
// reader. Possibly filter by session.
func ParseCmdLog(reader LineReader, arg ParseArgs) (err error) {

	var re *regexp.Regexp = nil
	if arg.Grep != "" {
		re = regexp.MustCompile(`\s+`)
		arg.Grep = re.ReplaceAllString(arg.Grep, ".*")
		re, err = regexp.Compile(arg.Grep)
		if err != nil {
			return ErrorLn("Failed to compile regexp \"", arg.Grep, "\": ", err)
		}
	}

	var since int64 = 0
	if arg.Since != "" {
		sincetm, err := time.ParseInLocation(timeFormat, arg.Since, time.Local)
		if err != nil {
			return ErrorLn("Parsing given since failed:", err)
		}
		since = sincetm.Unix()
	}

	// The format for the report structure:
	// for each element: timestring, session, command, [cwd]
	// If the strings in the element are empty, it has been filtered out
	report := make([][]string, initialReportLen)
	index := 0
	reportLock := sync.RWMutex{}

	type reportLine struct {
		line  string
		index int
	}
	jobs := make(chan reportLine, 10)

	wg := sync.WaitGroup{}
	worker := func(jobs <-chan reportLine) {
		for rl := range jobs {
			reportLock.RLock()
			report[rl.index] = make([]string, 4)
			ParseCmdLogLineNoAlloc(string(rl.line), arg.Session,
				since, re, &report[rl.index])
			reportLock.RUnlock()
		}
		wg.Done()
	}
	for i := 0; i < runtime.NumCPU()*2; i++ {
		wg.Add(1)
		go worker(jobs)
	}

	for {
		line, err := reader.ReadLine()

		if err == io.EOF {
			break
		}
		if err != nil {
			return ErrorLn("Error reading log: ", err)
		}
		if index >= len(report) {
			reportLock.Lock()
			report = append(report, []string{})
			reportLock.Unlock()
		}
		jobs <- reportLine{line, index}
		index = index + 1
	}
	close(jobs)
	wg.Wait()

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
		if report[pos] != nil && len(report[pos]) != 4 {
			debugTrace("report line is not nil, but len", len(report[pos]),
				"!= 4 pos", pos, "reportlen", reportlen)
		}

		if len(report[pos]) == 4 && report[pos][0] != "" {
			line := ""
			if arg.Session == "" {
				line = report[pos][1] + " "
			}
			line = line + report[pos][0]
			if arg.Pwd {
				line = line + "\t" + report[pos][3]
			}
			line = line + "\t" + report[pos][2]
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
