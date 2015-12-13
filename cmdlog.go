package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

// MajorVersion is the hard coded major version as opposed to the version
// provided from command line.
var MajorVersion = "0"

var (
	cmdlogFile = os.ExpandEnv("${HOME}/.cmdlog")
	version   = "Undefined"
	timestamp = "Undefined"
	homeDir   = os.Getenv("HOME")
	filters   = []string{
		" *ls? -[thlroa]*$",
		" *l[shla]*$",
	}
)

const (
	day              time.Duration = time.Hour * 24
	initialReportLen               = 16864
)

func mainLog(c *cli.Context) {
	args := c.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: source and argument are required.")
		os.Exit(1)
	}

	cmd := strings.Join(args[1:], " ")

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

	tm := time.Now()
	name := c.GlobalString("file")
	fp, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file \"%s\" for writing: %v",
			name, err)
		os.Exit(1)
	}
	fmt.Fprintf(fp, "%d\t%s\t%s\n", tm.Unix(), args[0], cmd)
	fp.Close()
}

var magnitudes = []struct {
	mag  time.Duration
	name string
}{
	{day, "day"},
	{time.Hour, "hour"},
	{time.Minute, "minute"},
	{time.Second, "second"},
}

// Converts a duration to a string according to magnitudes above
func formatRelativeTime(diff time.Duration) string {
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

// Formats the given timestring (UNIX time) to human readable string
func formatTime(timestr string) string {
	timeint, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		log.Printf("could not parse %s to integer: %v", timestr, err)
		return timestr
	}
	tm := time.Unix(timeint, 0)
	diff := time.Since(tm)

	if diff.Hours() < 24.0*7 {
		return formatRelativeTime(diff)
	}

	return tm.Format("2006-01-02T15:04:05")
}

// Prepares a single line for display
func parseCmdLogLine(line string, session string, regex *regexp.Regexp,
	out *[]string) {
	items := strings.SplitN(line, "\t", 3)

	// If session filtering is used and session does not match
	if session != "" && session != items[1] {
		return
	}

	// If regex is given and it does not match
	if regex != nil && !regex.MatchString(items[2]) {
		return
	}

	items[0] = formatTime(items[0])
	copy(*out, items)
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

// Add working directories to the report
func addPwdsToReport(report *[][]string) {
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

// Extendable list of arguments for the parseCmdLog function
type parseArgs struct {
	session string
	grep    string
	pwd     bool
	reverse bool
}

// Parses and prints out the command log from given reader. Possibly filter by
// session.
func parseCmdLog(input io.Reader, arg parseArgs) {
	scanner := bufio.NewScanner(input)

	var re *regexp.Regexp
	re = nil
	if arg.grep != "" {
		var err error
		re, err = regexp.Compile(arg.grep)
		if err != nil {
			log.Fatal("Failed to compile regexp ", arg.grep, err)
		}
	}

	// The format for the report structure:
	// for each element: timestring, session, command, [cwd]
	// If the strings in the element are empty, it has been filtered out
	report := make([][]string, initialReportLen)
	index := 0

	for scanner.Scan() {
		if index >= len(report) {
			report = append(report, []string{})
		}
		report[index] = make([]string, 4)
		parseCmdLogLine(scanner.Text(), arg.session, re, &report[index])
		index = index + 1
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading log:", err)
	}

	if arg.pwd {
		addPwdsToReport(&report)
	}

	// Print out the report
	reportlen := len(report) - 1
	for idx := range report {
		pos := idx
		if arg.reverse {
			pos = reportlen - idx
		}
		if report[pos] != nil && report[pos][0] != "" {
			line := ""
			if arg.session == "" {
				line = report[pos][1] + " "
			}
			line = line + report[pos][0]
			if arg.pwd {
				line = line + "\t" + report[pos][3]
			}
			line = line + "\t" + report[pos][2]
			fmt.Println(line)
		}
	}
}

func mainReport(c *cli.Context) {
	fp := os.Stdin
	name := c.GlobalString("file")
	if strings.Compare(name, "-") != 0 {
		var err error
		fp, err = os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()
	}
	arg := parseArgs{
		session: c.String("session"),
		grep:    c.String("grep"),
		pwd:     c.Bool("pwd"),
		reverse: c.Bool("reverse"),
	}
	parseCmdLog(fp, arg)
}

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "Command logging"
	app.Version = fmt.Sprintf("%s-%s", MajorVersion, version)
	app.Copyright = "MIT"
	app.Authors = []cli.Author{
		{
			Name:  "Kalle Kankare",
			Email: "kalle.kankare@iki.fi",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "file, f",
			Value:  cmdlogFile,
			Usage:  "Read commands from FILE",
			EnvVar: "CMDLOG_FILE",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "log",
			Usage:  "Log a new item. \n\n   Arguments: <source> <command> [args ...]",
			Action: mainLog,
		},
		{
			Name:   "report",
			Usage:  "Generate report from the command log",
			Action: mainReport,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "pwd",
					Usage:  "Print also the current directory where the command was run.",
					EnvVar: "CMDLOG_PWD",
				},
				cli.StringFlag{
					Name:   "session",
					Usage:  "Report lists commands of the given SESSION.",
					EnvVar: "CMDLOG_SESSION",
				},
				cli.StringFlag{
					Name:   "since, d",
					Usage:  "Display command starting from SINCE.",
					EnvVar: "CMDLOG_SINCE",
				},
				cli.BoolFlag{
					Name:   "reverse, r",
					Usage:  "Display command in reverse",
					EnvVar: "CMDLOG_REVERSE",
				},
				cli.StringFlag{
					Name:   "grep",
					Usage:  "Grep for a command",
					EnvVar: "CMDLOG_GREP",
				},
			},
		},
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s %v\nbuilt %v\n", app.Name,
			app.Version, timestamp)
	}

	app.Run(os.Args)
}
