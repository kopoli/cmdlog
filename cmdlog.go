package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	cmdlib "./lib"
	"github.com/kopoli/go-util"
)

// MajorVersion is the hard coded major version as opposed to the version
// provided from command line.
var MajorVersion = "1"

var (
	cmdlogFile = os.ExpandEnv("${HOME}/.cmdlog")
	version    = "Undefined"
	timestamp  = "Undefined"
	exitValue int = 0
)

func checkErr(err error, message string, arg ...string) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %s%s. (error: %s)\n", message,
		strings.Join(arg, " "), err)

	// Exit goroutine and run all deferrals
	exitValue = 1
	runtime.Goexit()
}

type profiler struct {
	cpuproffile string
	memproffile string
}

func createProfileFile(outfile string) *os.File {
	fp, err := os.Create(outfile)
	checkErr(err, "Could not create profile file", outfile)
	return fp
}

func setupProfiler(opts util.Options) profiler {
	cpuproffile := opts.Get("profile-cpu-file", "")
	memproffile := opts.Get("profile-mem-file", "")
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

func main() {
	opts := util.NewOptions()
	opts.Set("program-version", version)
	opts.Set("program-timestamp", timestamp)
	opts.Set("cmdlog-file", cmdlogFile)

	// In the last deferred function, exit the program with given code
	defer func() {
		os.Exit(exitValue)
	}()

	_, err := cmdlib.Cli(opts, os.Args)
	checkErr(err, "Parsing command line failed")

	op := opts.Get("operation", "")
	cmdlogFile = opts.Get("cmdlog-file", "jeje")

	p := setupProfiler(opts)
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
