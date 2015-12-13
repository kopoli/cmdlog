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
	version   = "Undefined"
	timestamp = "Undefined"
)

func main() {
	app := cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "Command logging"
	app.Version = fmt.Sprintf("%s-%s", MajorVersion, version)
	app.Commands = []cli.Command{
		{
			Name:  "log",
			Usage: "Log a new item",
		},
		{
			Name:  "report",
			Usage: "Generate report",
		},
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s version %v built %v\n", app.Name,
			app.Version, timestamp)
	}

	app.Run(os.Args)
}
