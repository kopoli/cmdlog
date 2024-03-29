package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/kopoli/appkit"
	cmdlib "github.com/kopoli/cmdlog/lib"
)

var (
	version     = "Undefined"
	timestamp   = "Undefined"
	buildGOOS   = "Undefined"
	buildGOARCH = "Undefined"

	cmdlogFile       = os.ExpandEnv("${HOME}/.cmdlog")
	cmdlogFilterFile = os.ExpandEnv("${HOME}/.cmdlog-filters")

	maximumLineLength = 160 * 1024
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
		err = pprof.WriteHeapProfile(fp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Could not write memory profile fil:", err)
			return
		}
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
	opts.Set("cmdlog-filter-file", cmdlogFilterFile)

	exitValue := 0

	// In the last deferred function, exit the program with given code
	defer func() {
		// Only exit properly if not panicing
		if e := recover(); e != nil {
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

	op := opts.Get("cmdline-command", "")
	if op == "done" {
		// No additional operations
		return
	}
	cmdlogFile = opts.Get("cmdlog-file", cmdlogFile)
	cmdlogFilterFile = opts.Get("cmdlog-filter-file", cmdlogFilterFile)
	log := cmdlib.CreateLog(cmdlogFile, cmdlogFilterFile)

	p, err := setupProfiler(opts)
	checkErr(err, "Could not create profile file")
	defer p.deleteProfiler()

	handleFilters := func() {
		// Save default filters if the filter file doesn't exist
		err = log.SaveDefaultFilters()
		checkErr(err, "Could not write default filters")

		// Load filters
		err = log.LoadFilters()
		if err != nil {
			// Failure of loading filters is not a fatal error
			fmt.Fprintf(os.Stderr,
				"Warning: Problems loading filters: %v", err)
		}
	}

	switch op {
	case "log":
		handleFilters()

		session := opts.Get("log-session", "<unknown>")
		args := opts.Get("log-args", "<unknown>")

		err = log.AppendLine(session, args)
		checkErr(err, "Could not print to log")
	case "filters":
		handleFilters()
		for i := range log.Filters {
			fmt.Println(log.Filters[i])
		}
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
			lr, err = cmdlib.NewReverseReader(fp, maximumLineLength)
			checkErr(err, "Creating a new reverse reader failed")
		} else {
			lr = cmdlib.NewBufferedReader(fp, maximumLineLength)
		}

		err = cmdlib.ParseCmdLog(lr, arg)
		checkErr(err, "Parsing the command log failed")
	default:
		err = fmt.Errorf("invalid command")
		checkErr(err, "Running cmdlog failed")
	}
}
