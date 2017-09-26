package cmdlib

import (
	"fmt"
	"runtime"
	"strings"

	cli "github.com/jawher/mow.cli"

	"github.com/kopoli/go-util"
)

func Cli(opts util.Options, argsin []string) (args []string, err error) {
	progName := opts.Get("program-name", "cmdlog")
	progVersion := opts.Get("program-version", "undefined")
	progTimestamp := opts.Get("program-timestamp", "undefined")

	app := cli.App(progName, "Command logging and reporting")

	app.Version("version v", fmt.Sprintf("%s: %s\nBuilt %v with: %s/%s for %s/%s",
		progName, progVersion, progTimestamp, runtime.Compiler,
		runtime.Version(), runtime.GOOS, runtime.GOARCH))

	app.Command("log", "Log a new command line", func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS] SOURCE COMMAND [ARGS...]"

		argSource := cmd.StringArg("SOURCE", "", "Command source indentifier")
		argCommand := cmd.StringArg("COMMAND", "", "Command")
		argArgs := cmd.StringsArg("ARGS", []string{}, "Command arguments")

		cmd.Action = func() {
			opts.Set("operation", "log")
			opts.Set("log-source", *argSource)
			opts.Set("log-args", *argCommand+" "+strings.Join(*argArgs, " "))
		}
	})

	app.Command("report", "Generate a report from the command log", func(cmd *cli.Cmd) {
		optPwd := cmd.BoolOpt("pwd", false,
			"Print also the current directory where the command was run")
		optSession := cmd.StringOpt("session", "",
			"List commands of the given session")
		optSince := cmd.StringOpt("d since", "",
			"Display commands starting from given date")
		optReverse := cmd.BoolOpt("r reverse", false,
			"Display commands in reverse")
		optGrep := cmd.StringOpt("grep", "",
			"Display commands matching given regular expression")

		cmd.Action = func() {
			opts.Set("operation", "report")
			if *optPwd {
				opts.Set("report-pwd", "t")
			}
			if *optReverse {
				opts.Set("report-reverse", "t")
			}
			opts.Set("report-session", *optSession)
			opts.Set("report-since", *optSince)
			opts.Set("report-grep", *optGrep)
		}
	})

	optCmdFile := app.StringOpt("f file", opts.Get("cmdlog-file", "cmdlog.debug"),
		"File name of the command log")
	optCpuProfile := app.StringOpt("profile cpuprofile", "",
		"File name to save CPU profile")
	optMemProfile := app.StringOpt("memprofile", "",
		"File name to save memory profile")

	app.After = func() {
		opts.Set("cmdlog-file", *optCmdFile)
		opts.Set("profile-cpu-file", *optCpuProfile)
		opts.Set("profile-mem-file", *optMemProfile)
	}

	err = app.Run(argsin)
	return
}
