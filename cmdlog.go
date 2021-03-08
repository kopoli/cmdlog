package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	cmdlib "github.com/kopoli/cmdlog/lib"
	"github.com/kopoli/appkit"
)

var (
	version     = "Undefined"
	timestamp   = "Undefined"
	buildGOOS   = "Undefined"
	buildGOARCH = "Undefined"

	cmdlogFile     = os.ExpandEnv("${HOME}/.cmdlog")
)

type profiler struct {
	cpuproffile string
	memproffile string
}

func setupProfiler(opts appkit.Options) (*profiler, error) {
	cpuproffile := opts.Get("profile-cpu-file", "")
	memproffile := opts.Get("profile-mem-file", "")
	ret := &profiler{
		cpuproffile: cpuproffile,
		memproffile: memproffile,
	}

	if cpuproffile != "" {
		fp, err := os.Create(cpuproffile)
		if err != nil {
			return nil, err
		}
		runtime.SetCPUProfileRate(1000)
		err = pprof.StartCPUProfile(fp)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (p *profiler) deleteProfiler() {
	if p.cpuproffile != "" {
		pprof.StopCPUProfile()
	}
	if p.memproffile != "" {
		fp, err := os.Create(p.memproffile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not create memory profile file:", err)
			return
		}
		defer fp.Close()
		pprof.WriteHeapProfile(fp)
	}
}

func main() {
	opts := appkit.NewOptions()
	opts.Set("program-name", os.Args[0])
	opts.Set("program-version", version)
	opts.Set("program-timestamp", timestamp)
	opts.Set("program-buildgoos", buildGOOS)
	opts.Set("program-buildgoarch", buildGOARCH)
	opts.Set("cmdlog-file", cmdlogFile)

	exitValue := 0

	// In the last deferred function, exit the program with given code
	defer func() {
		// Only exit properly if not panicing
		if e:= recover(); e != nil {
			panic(e)
		} else {
			os.Exit(exitValue)
		}
	}()

	checkErr := func(err error, message string, arg ...string) {
		if err == nil {
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %s%s. (error: %s)\n", message,
			strings.Join(arg, " "), err)

		// Exit goroutine and run all deferrals
		exitValue = 1
		runtime.Goexit()
	}

	err := cmdlib.Cli(opts, os.Args[1:])
	checkErr(err, "Parsing command line failed")

	op := opts.Get("operation", "")
	if op == "" {
		// No additional operations
		return
	}
	cmdlogFile = opts.Get("cmdlog-file", "jeje")

	p, err := setupProfiler(opts)
	checkErr(err, "Could not create profile file")
	defer p.deleteProfiler()

	switch op {
	case "log":
		fp, err := os.OpenFile(cmdlogFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		checkErr(err, "Could not open file \"",
			cmdlogFile, "\" for writing", cmdlogFile)
		defer fp.Close()

		source := opts.Get("log-source", "<unknown>")
		args := opts.Get("log-args", "<unknown>")

		err = cmdlib.AppendLine(fp, source, args)
		checkErr(err, "Could not print to log")
	case "report":
		arg := cmdlib.ParseArgs{
			Session: opts.Get("report-session", ""),
			Since:   opts.Get("report-since", ""),
			Grep:    opts.Get("report-grep", ""),
			Pwd:     opts.IsSet("report-pwd"),
			Output:  os.Stdout,
		}
		fp := os.Stdin
		if strings.Compare(cmdlogFile, "-") != 0 {
			fp, err = os.Open(cmdlogFile)
			checkErr(err, "Could not open", cmdlogFile, "for reading.")
			defer fp.Close()
		}
		var lr cmdlib.LineReader
		if opts.IsSet("report-reverse") {
			lr, err = cmdlib.NewReverseReader(fp)
			checkErr(err, "Creating a new reverse reader failed")
		} else {
			lr = cmdlib.NewBufferedReader(fp)
		}

		err = cmdlib.ParseCmdLog(lr, arg)
		checkErr(err, "Parsing the command log failed")
	default:
		err = errors.New(fmt.Sprintf("Invalid command: %s", op))
		checkErr(err, "Running cmdlog failed")
	}
}
