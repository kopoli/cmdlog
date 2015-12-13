package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
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
	cmdlogFile = os.ExpandEnv(".cmdlog")
	// cmdlogFile = os.ExpandEnv("${HOME}/.cmdlog")
	version   = "Undefined"
	timestamp = "Undefined"
)

func mainLog(c *cli.Context) {
	// TODO Implement this
	fmt.Println("Adding a log entry is currently unsupported.")
}

const (
	day time.Duration = time.Hour * 24
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
			ret = ret + strconv.Itoa(int(count)) + " " + mag.name
			if count > 1 {
				ret = ret + "s"
			}
			ret = ret + " "
		}
	}

	return strings.TrimSuffix(ret, " ")
}

// Formats the given timestring (UNIX time) to human readable string
func formatTime(timestr string) string {
	timeint, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		log.Printf("could not parse %s to integer", timestr)
		return timestr
	}
	tm := time.Unix(timeint, 0)
	diff := time.Since(tm)

	if diff.Hours() < 24.0*7 {
		return formatRelativeTime(diff)
	}

	tm.Round(time.Second)
	return tm.String()
}

// Prepares a single line for display
func parseCmdLogLine(line string, session string, regex *regexp.Regexp) []string {
	items := strings.SplitN(line, "\t", 3)

	// If session filtering is used and session does not match
	if session != "" && session != items[1] {
		return nil
	}

	if regex != nil {
		if ! regex.MatchString(items[2]) {
			return nil
		}
	}

	items[0] = formatTime(items[0])

	ret := items[1] + " " + items[0] + "\t" + items[2]

	fmt.Println(ret)
	return items
}

// Parses and prints out the command log from given reader. Possibly filter by
// session.
func parseCmdLog(input io.Reader, session string, grep string) {
	scanner := bufio.NewScanner(input)

	var re *regexp.Regexp
	re = nil
	if grep != "" {
		var err error
		re, err = regexp.Compile(grep)
		if err != nil {
			log.Fatal("Failed to compile regexp ", grep, err)
		}
	}

	for scanner.Scan() {
		parseCmdLogLine(scanner.Text(), session, re)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading log:", err)
	}
}

func mainReport(c *cli.Context) {
	fp := os.Stdin
	name := c.GlobalString("file")
	fmt.Println(name)
	if strings.Compare(name, "-") != 0 {
		var err error
		fp, err = os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()
	}
	fmt.Println(c.String("session"))
	parseCmdLog(fp, c.String("session"), c.String("grep"))
}

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "Command logging"
	app.Version = fmt.Sprintf("%s-%s", MajorVersion, version)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "file",
			Value:  cmdlogFile,
			Usage:  "Read commands from FILE",
			EnvVar: "CMDLOG_FILE",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "log",
			Usage:  "Log a new item",
			Action: mainLog,
		},
		{
			Name:   "report",
			Usage:  "Generate report from the commanad log",
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
