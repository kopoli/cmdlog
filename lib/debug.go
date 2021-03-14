package cmdlib

import (
	"fmt"
	"os"
	"time"
)

var (
	debugLogFile = os.Getenv("CMDLOG_DEBUG_TRACE")
)

func DebugTrace(args ...interface{}) {
	if debugLogFile == "" {
		return
	}
	f, err := os.OpenFile(debugLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(fmt.Sprint("Debug file opening failed:", err))
	}
	defer f.Close()

	args = append(args, time.Now().String())

	_, err = fmt.Fprintln(f, args...)
	if err != nil {
		panic(fmt.Sprint("Appending to debug file failed", err))
	}
}
