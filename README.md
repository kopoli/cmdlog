# Cmdlog

Fleshed out command history logging.

## Installation

```
$ go get github.com/kopoli/cmdlog
```

## Description

This is a shell history logging similar to the `~/.bash_history` and `~/.zsh_history` files.
It logs the following items for each command line arguments

- Session identifier in which the command was executed.
- Timestamp when the command was executed.

The purpose of this is to provide a filter-like view of the command history and retrieve the commands to the current command line.
The filtering view is implemented using https://github.com/kopoli/thelm

## Demo

![Demo running cmdlog](./_example/cmdlog-demo.svg)

To set up running the demo for yourself do the following:

Install `zsh` via a package manager.

Run the following commands:

```
# Clone this repository
git clone github.com/kopoli/cmdlog
cd cmdlog

# Install the cmdlog binary
go install

# Install the thelm program for filtering
go get https://github.com/kopoli/thelm
```

Start the example with the following commands:
```
cd _example
./run-example.sh
```

(The above demo is recorded using https://github.com/nbedos/termtosvg)

## Usage

```
$ cmdlog -help

cmdlog: Command logging and reporting.

Usage: cmdlog [OPTIONS] <COMMAND>

Commands:
  log      -  Log a new command line
  report   -  Generate a report from the command log
  filters  -  Print log line filters

Options:
  -file string
    	File name of the command log ($CMDLOG_FILE) (default "$HOME/.cmdlog")
  -filter string
    	File name of the command line filter file ($CMDLOG_FILTERS) (default "$HOME/.cmdlog-filters")
  -memprofile string
    	File name to save memory profile ($CMDLOG_MEMPROFILE)
  -profile string
    	File name to save CPU profile ($CMDLOG_CPUPROFILE)
  -v	Display version
  -version
    	Display version
```

### Log

```
$ cmdlog log -help

Command: log [OPTIONS] SESSION ARGS[...]

Log a new command line

Parameters:
  SESSION   Command session identifier
  ARGS      Command line arguments
```

Example:
```
cmdlog log shell-session-1 go build
```

results in the following to be inserted into `~/.cmdlog`:
```
1617900929	shell-session-1	go build
```

#### Report

```
$ cmdlog report -help

Command: report

Generate a report from the command log

Options:
  -grep string
    	Display commands matching given regular expression
  -pwd
    	Print also the current directory where the command was run
  -reverse
    	Display commands in reverse
  -session string
    	List commands of the given session
  -since string
    	Display commands starting from given date
```

Display commands from the command log.

Example:
```
$ cmdlog -grep build
shell-session-1 8s ago	go build
```

## License

MIT license
