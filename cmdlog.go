package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

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
)

type profiler struct {
	cpuproffile string
	memproffile string
}

func createProfileFile(outfile string) *os.File {
	fp, err := os.Create(outfile)
	if err != nil {
		cmdlib.Panicln("Could not create profile file", outfile, err)
	}
	return fp
}

func createProfiler(cpuproffile, memproffile string) profiler {
	ret := profiler{
		cpuproffile: cpuproffile,
		memproffile: memproffile,
	}

	if cpuproffile != "" {
		fp := createProfileFile(cpuproffile)
		runtime.SetCPUProfileRate(1000)
		pprof.StartCPUProfile(fp)
	}
	return ret
}

func (p *profiler) deleteProfiler() {
	if p.cpuproffile != "" {
		pprof.StopCPUProfile()
	}
	if p.memproffile != "" {
		fp := createProfileFile(p.memproffile)
		defer fp.Close()
		pprof.WriteHeapProfile(fp)
	}
}

func mainLog(c *cli.Context) {
	defer cmdlib.Recover()
	args := c.Args()
	if len(args) < 2 {
		cmdlib.Panicln("Source and argument are required.")
	}

	logfile := c.GlobalString("file")
	fp, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		cmdlib.Panicln("Could not open file \"", logfile, "\" for writing:",
			logfile, err)
	}
	cmdlib.AppendLine(fp, args[0], args[1:])
	fp.Close()
}

func mainReport(c *cli.Context) {
	defer cmdlib.Recover()
	proffile := c.GlobalString("profile")
	var prof profiler
	if proffile != "" {
		prof = createProfiler(c.GlobalString("profile"), c.GlobalString("memprofile"))
		defer prof.deleteProfiler()
	}

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
		Output: os.Stdout,
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
		cli.StringFlag{
			Name:   "profile",
			Usage:  "Create a CPU profile of the runtime",
			EnvVar: "CMDLOG_PROFILE",
		},
		cli.StringFlag{
			Name:   "memprofile",
			Usage:  "Create a memory profile of the runtime",
			EnvVar: "CMDLOG_MEMPROFILE",
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
		fmt.Fprintf(c.App.Writer, "%s %v\nbuilt %v with Go %v\n",
			app.Name, app.Version, timestamp, runtime.Version())
	}

	err := app.Run(os.Args)
	if err != nil {
		cmdlib.Panicln("Running the application: ", err)
		defer cmdlib.Recover()
	}
}
