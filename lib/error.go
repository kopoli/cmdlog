package cmdlib

import (
	"errors"
	"fmt"
)

func Warningln(v ...interface{}) {
	// do nothing
}

func ErrorLn(v ...interface{}) error {
	return errors.New(fmt.Sprint(v...))
}
