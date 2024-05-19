package zygo

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

var Verbose bool = true // set to true to debug
var Working bool        // currently under investigation

// var V = VPrintf
var W = WPrintf
var Q = func(quietly_ignored ...interface{}) {} // quiet

var vv = VPrintf2

// P is a shortcut for a call to fmt.Printf that implicitly starts
// and ends its message with a newline.
func P(format string, stuff ...interface{}) {
	fmt.Printf("\n "+format+"\n", stuff...)
}

// get timestamp for logging purposes
func ts() string {
	return time.Now().Format("2006-01-02 15:04:05.999 -0700 MST")
}

// so we can multi write easily, use our own printf
var OurStdout io.Writer = os.Stdout

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(OurStdout, format, a...)
}

var tsPrintfMut sync.Mutex

// time-stamped printf
func TSPrintf(format string, a ...interface{}) {
	tsPrintfMut.Lock()
	Printf("\n%s %s ", FileLine(3), ts())
	Printf(format+"\n", a...)
	tsPrintfMut.Unlock()
}

func VPrintf2(format string, a ...interface{}) {
	if Verbose {
		TSPrintf(format+"\n", a...)
	}
}

func WPrintf(format string, a ...interface{}) {
	if Working {
		TSPrintf(format+"\n", a...)
	}
}

func FileLine(depth int) string {
	_, fileName, fileLine, ok := runtime.Caller(depth)
	var s string
	if ok {
		s = fmt.Sprintf("%s:%d", path.Base(fileName), fileLine)
	} else {
		s = ""
	}
	return s
}

func Caller(upStack int) string {
	// elide ourself and runtime.Callers
	target := upStack + 2

	pc := make([]uintptr, target+2)
	n := runtime.Callers(0, pc)

	f := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(pc[:n])
		for i := 0; i <= target; i++ {
			contender, more := frames.Next()
			if i == target {
				f = contender
			}
			if !more {
				break
			}
		}
	}
	return f.Function
}
