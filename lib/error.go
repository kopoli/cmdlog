package cmdlib

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func Warningln(v ...interface{}) {
	// do nothing
}

func ErrorLn(v ...interface{}) error {
	return errors.New(fmt.Sprint(v...))
}

func FatalErr(err error, message string, arg ...string) {
	msg := ""
	if err != nil {
		msg = fmt.Sprintf(" (error: %s)", err)
	}
	fmt.Fprintf(os.Stderr, "Error: %s%s.%s\n", message, strings.Join(arg, " "), msg)
	os.Exit(1)
}
