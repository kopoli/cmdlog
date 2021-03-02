package cmdlib

import (
	"flag"
	"fmt"
	"strings"

	"github.com/kopoli/appkit"
)

func Cli(opts appkit.Options, argsin []string) error {
	base := appkit.NewCommand(nil, "", "Command logging and reporting")
	optVersion := base.Flags.Bool("version", false, "Display version")

	optCmdFile := base.Flags.String("file", opts.Get("cmdlog-file", "cmdlogs.debug"),
		"File name of the command log")
	optCpuProfile := base.Flags.String("profile", "", "File name to save CPU profile")
	optMemProfile := base.Flags.String("memprofile", "", "File name to save memory profile")

	log := appkit.NewCommand(base, "log", "Log a new command line")

	log.Flags.Usage = func() {
		out := log.Flags.Output()
		fmt.Fprintf(out, "Command: log [OPTIONS] SOURCE ARGS[...]\n\n"+
			"  SOURCE    Command source identifier\n"+
			"  ARGS      Command line arguments\n")
	}

	report := appkit.NewCommand(base, "report", "Generate a report from the command log")
	optPwd := report.Flags.Bool("pwd", false,
		"Print also the current directory where the command was run")
	optSession := report.Flags.String("session", "",
		"List commands of the given session")
	optSince := report.Flags.String("since", "",
		"Display commands starting from given date")
	optReverse := report.Flags.Bool("reverse", false,
		"Display commands in reverse")
	optGrep := report.Flags.String("grep", "",
		"Display commands matching given regular expression")

	err := base.Parse(argsin, opts)
	if err == flag.ErrHelp || *optVersion {
		if *optVersion {
			fmt.Println(appkit.VersionString(opts))
		}
		opts.Set("operation", "")
		return nil
	}

	opts.Set("cmdlog-file", *optCmdFile)
	opts.Set("profile-cpu-file", *optCpuProfile)
	opts.Set("profile-mem-file", *optMemProfile)

	cmd := opts.Get("cmdline-command", "")
	switch cmd {
	case "log":
		args := appkit.SplitArguments(opts.Get("cmdline-args", ""))
		if len(args) < 2 {
			return fmt.Errorf("Invalid arguments for log: %s",
				strings.Join(args, " "))
		}
		opts.Set("operation", "log")
		opts.Set("log-source", args[0])
		opts.Set("log-args", strings.Join(args[1:], " "))
	case "report":
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

	return nil
}
