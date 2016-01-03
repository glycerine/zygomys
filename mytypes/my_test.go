package mytypes

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test001TypeReflectionOnMytypeEvent(t *testing.T) {

	cv.Convey("from glisp we should be able to create the known Go struct, Event{}", t, func() {
		lisp := `(event id:"abc" game:"craps" win:true bet-amount:"50.00")`
		cv.So()
	})
}
