package logging

import (
	"fmt"
	"os"
)

// DebugEnabled returns true if debug mode is enabled via TT_DEBUG environment variable
func DebugEnabled() bool {
	return os.Getenv("TT_DEBUG") != ""
}

// Debugf prints a formatted debug message only if debug mode is enabled
func Debugf(format string, args ...interface{}) {
	if DebugEnabled() {
		fmt.Printf(format, args...)
	}
}

// Debugln prints a debug message followed by a newline only if debug mode is enabled
func Debugln(args ...interface{}) {
	if DebugEnabled() {
		fmt.Println(args...)
	}
}
