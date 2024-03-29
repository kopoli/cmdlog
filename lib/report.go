package cmdlib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	day time.Duration = time.Hour * 24
)

var (
	homeDir    = os.Getenv("HOME")
	timeFormat = "2006-01-02T15:04:05"
)

var magnitudes = []struct {
	mag time.Duration
	// If the given magnitude is above this, do not continue listing
	// smaller magnitudes.
	minimumBreak int64
	name         string
}{
	{day, 2, "d"},
	{time.Hour, 12, "h"},
	{time.Minute, 30, "m"},
	{time.Second, 60, "s"},
}

// FormatRelativeTime converts a duration to a string according to magnitudes above
func FormatRelativeTime(diff time.Duration) string {
	var ret string

	if diff.Seconds() < 1.0 {
		return "Just now"
	}

	for _, mag := range magnitudes {
		count := diff / mag.mag
		diff %= mag.mag
		if count > 0 {
			ret = ret + strconv.FormatInt(int64(count), 10) + mag.name + " "
		}
		if int64(count) > mag.minimumBreak {
			break
		}
	}

	return ret + "ago"
}

// FormatTime formats the given timestring (UNIX time) to human readable string
func FormatTime(timeint int64, now time.Time) string {
	tm := time.Unix(timeint, 0)
	diff := now.Sub(tm)

	if diff.Hours() < 24.0*7 {
		return FormatRelativeTime(diff)
	}

	return tm.Format(timeFormat)
}

// ParseCmdLogLineNoAlloc prepares a single line without unnecessary allocation.
func ParseCmdLogLineNoAlloc(line string, session string, since int64, now time.Time, regex *regexp.Regexp,
	out *[]string) {
	var pos [2]int
	start := 0
	for i := range pos {
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
	switch {
	case err != nil:
		(*out)[0] = "<invalid>"
	case timeint < since:
		return
	default:
		(*out)[0] = FormatTime(timeint, now)
	}

	(*out)[1] = line[pos[0] : pos[1]-1]
	(*out)[2] = line[pos[1]:]
}

type controlArgs struct {
	JobCount             int
	Now                  time.Time
	BufferLineCount      int
	CompletionBufferSize int
	ReportLen            int
}

func defaultControlArgs() controlArgs {
	return controlArgs{
		JobCount:             runtime.NumCPU(),
		Now:                  time.Now(),
		BufferLineCount:      24,
		CompletionBufferSize: 1024,
		ReportLen:            1024 * 128,
	}
}

func (ca *controlArgs) FillDefault() {
	def := defaultControlArgs()
	if ca.JobCount == 0 {
		ca.JobCount = def.JobCount
	}
	if ca.Now == (time.Time{}) {
		ca.Now = def.Now
	}
	if ca.BufferLineCount == 0 {
		ca.BufferLineCount = def.BufferLineCount
	}
	if ca.CompletionBufferSize == 0 {
		ca.CompletionBufferSize = def.CompletionBufferSize
	}
	if ca.ReportLen == 0 {
		ca.ReportLen = def.ReportLen
	}
}

// ParseArgs is extendable list of arguments for the parseCmdLog function
type ParseArgs struct {
	Session string
	Since   string
	Grep    string
	Pwd     bool
	Control controlArgs
	Output  io.Writer
}

// ParseCmdLog Parses and prints out the command log from given
// reader. Possibly filter by session.
func ParseCmdLog(reader LineReader, arg ParseArgs) (err error) {
	arg.Control.FillDefault()

	var filterRe *regexp.Regexp
	if arg.Grep != "" {
		filterRe = regexp.MustCompile(`\s+`)
		arg.Grep = filterRe.ReplaceAllString(arg.Grep, ".*")
		filterRe, err = regexp.Compile(arg.Grep)
		if err != nil {
			return fmt.Errorf("failed to compile regexp \"%s\": %s", arg.Grep, err)
		}
	}

	var since int64
	if arg.Since != "" {
		sincetm, err := time.ParseInLocation(timeFormat, arg.Since, time.Local)
		if err != nil {
			return fmt.Errorf("parsing given since failed: %s", err)
		}
		since = sincetm.Unix()
	}

	out := NewBufferedWriter(arg.Output, arg.Control.BufferLineCount)

	// The format for the report structure:
	// for each element: timestring, session, command, [cwd]
	// If the strings in the element are empty, it has been filtered out
	// Tried to make this a [][4]string, but that was half the speed that
	// this currently is.
	report := make([][]string, arg.Control.ReportLen)
	index := 0
	reportLock := sync.RWMutex{}

	type reportLine struct {
		line  string
		index int
	}
	jobs := make(chan reportLine, arg.Control.JobCount)
	completions := make(chan int, arg.Control.JobCount)

	wg := sync.WaitGroup{}

	// Parses the report line strings to the report array
	worker := func(jobs <-chan reportLine, completions chan<- int) {
		for rl := range jobs {
			reportLock.RLock()
			report[rl.index] = make([]string, 4)
			ParseCmdLogLineNoAlloc(rl.line, arg.Session,
				since, arg.Control.Now, filterRe, &report[rl.index])
			reportLock.RUnlock()
			completions <- rl.index
		}
		wg.Done()
	}

	for i := 0; i < arg.Control.JobCount*2; i++ {
		wg.Add(1)
		go worker(jobs, completions)
	}

	// Print a single report line
	printLine := func(pos int) {
		reportLock.RLock()
		if len(report[pos]) == 4 && report[pos][0] != "" {
			// A stringbuilder was tried here, but that allocated
			// 3MB more memory
			line := ""
			if arg.Session == "" {
				line = report[pos][1] + " "
			}
			line += report[pos][0]
			if arg.Pwd {
				line = line + "\t" + report[pos][3]
			}
			line = line + "\t" + report[pos][2]
			_, _ = out.Write([]byte(line))
		}
		reportLock.RUnlock()
	}

	printWg := sync.WaitGroup{}

	// Print the whole report as it gets parsed
	printer := func(completions <-chan int) {
		complete := make([]int, 0, arg.Control.CompletionBufferSize)
		firstNotPrinted := 0

		for idx := range completions {
			complete = append(complete, idx)
			sort.Ints(complete)

			var limit = firstNotPrinted
			var next = len(complete)

			// Get the number of sequential items that can be printed
			for i := range complete {
				if limit != complete[i] {
					next = i
					break
				}
				limit++
			}

			if limit != firstNotPrinted {
				for i := firstNotPrinted; i < limit; i++ {
					printLine(i)
				}

				complete = complete[next:]
				firstNotPrinted = limit
			}
		}
		printWg.Done()
	}

	// If PWD printing is enabled, it needs to be done after parsing
	printWg.Add(1)
	if arg.Pwd {
		// Drain the channel
		go func(completions <-chan int) {
			for range completions {
			}
			printWg.Done()
		}(completions)
	} else {
		go printer(completions)
	}

	// Read lines from the log
	for {
		line, err := reader.ReadLine()

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading log: %v", err)
		}
		if index >= cap(report)-1 {
			reportLock.Lock()

			// Allocate to capacity
			report = append(report, []string{})
			report = append(report, make([][]string, cap(report)-len(report))...)
			reportLock.Unlock()
		}
		jobs <- reportLine{line, index}
		index++
	}
	close(jobs)
	wg.Wait()
	close(completions)
	printWg.Wait()

	if arg.Pwd {
		AddPwdsToReport(&report)
		for idx := range report {
			printLine(idx)
		}
	}

	return out.Close()
}

// Heuristic to determine the current directory
func determineDirectory(previous string, cmd string) string {
	ret := ""

	switch {
	case cmd == "s":
		ret = ".."
	case strings.HasPrefix(cmd, "cd"):
		parts := strings.SplitN(cmd, " ", 2)
		if len(parts) == 1 {
			ret = homeDir
		} else {
			ret = parts[1]
		}
	case strings.HasPrefix(cmd, "Started shell session: "):
		parts := strings.SplitN(cmd, " ", 4)
		ret = parts[3]
	}

	if !filepath.IsAbs(ret) {
		ret = filepath.Join(previous, ret)
	}

	return strings.TrimSpace(filepath.Clean(ret))
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
