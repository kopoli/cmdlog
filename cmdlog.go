package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	cmdlib "./lib"
	"github.com/codegangsta/cli"
)

// MajorVersion is the hard coded major version as opposed to the version
// provided from command line.
var MajorVersion = "1"

var (
	cmdlogFile = os.ExpandEnv("${HOME}/.cmdlog")
	version    = "Undefined"
	timestamp  = "Undefined"
	filters    = []string{
		" *ls? -[thlroa]*$",
		" *l[shla]*$",
	}
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
	arg := cmdlib.ParseArgs{
		Session: c.String("session"),
		Grep:    c.String("grep"),
		Pwd:     c.Bool("pwd"),
		Reverse: c.Bool("reverse"),
	}
	cmdlib.ParseCmdLog(fp, arg)
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
