package zygo

import (
	"fmt"
	"time"
)

var Verbose bool = true // set to true to debug
var Working bool        // currently under investigation

var V = VPrintf
var W = WPrintf
var Q = func(quietly_ignored ...interface{}) {} // quiet

// P is a shortcut for a call to fmt.Printf that implicitly starts
// and ends its message with a newline.
func P(format string, stuff ...interface{}) {
	fmt.Printf("\n "+format+"\n", stuff...)
}

// get timestamp for logging purposes
func ts() string {
	return time.Now().Format("2006-01-02 15:04:05.999 -0700 MST")
}

// time-stamped printf
func TSPrintf(format string, a ...interface{}) {
	fmt.Printf("%s ", ts())
	fmt.Printf(format, a...)
}

func VPrintf(format string, a ...interface{}) {
	if Verbose {
		TSPrintf(format, a...)
	}
}

func WPrintf(format string, a ...interface{}) {
	if Working {
		TSPrintf(format, a...)
	}
}
