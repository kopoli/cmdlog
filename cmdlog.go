package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

// MajorVersion is the hard coded major version as opposed to the version
// provided from command line.
var MajorVersion = "0"

var (
	cmdlogFile = os.ExpandEnv("${HOME}/.cmdlog")
	version    = "Undefined"
	timestamp  = "Undefined"
)

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
			Name:  "log",
			Usage: "Log a new item",
		},
		{
			Name:  "report",
			Usage: "Generate report from the commanad log",
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
			},
		},
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s %v\n built %v\n", app.Name,
			app.Version, timestamp)
	}

	app.Run(os.Args)
}
