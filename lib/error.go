package cmdlib

import (
	"fmt"
	"os"
)

func Warningln(v ...interface{}) {
	// do nothing
}

func Panicln(v ...interface{}) {
	s := fmt.Sprint(v...)
	panic(s)
}

func Recover() {
	if r := recover(); r != nil {
		fmt.Fprintln(os.Stderr, "Error:", r)
		os.Exit(1)
	}
}
