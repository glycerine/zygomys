package zygo

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

func Test400SandboxFunctions(t *testing.T) {

	cv.Convey(`Given that the developer wishes to sandbox the Zygo interpreter when embedding it in their program, the NewGlispSandbox() function should return an interpreter that cannot call system/filesystem functions`, t, func() {

		sysFuncs := SystemFunctions()
		sandSafeFuncs := SandboxSafeFunctions()
		{
			env := NewGlispSandbox()

			// no system functions should pass
			for name := range sysFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? '%s)", name))
				cv.So(res, cv.ShouldResemble, SexpBool{Val: false})
				cv.So(err, cv.ShouldResemble, nil)
			}

			// all sandSafeFuncs should be fine
			for name := range sandSafeFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? '%s)", name))
				switch y := res.(type) {
				case SexpSentinel:
					P("'%s' wasn't defined but should be; defined? returned '%s'", name, y.SexpString())
				case SexpBool:
					cv.So(res, cv.ShouldResemble, SexpBool{Val: true})
				}
				cv.So(err, cv.ShouldEqual, nil)
			}
		}

		{
			fmt.Printf("\n and all functions should be reachable from a non-sandboxed environment.\n")
			env := NewGlisp()
			for name := range sysFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? '%s)", name))
				cv.So(res, cv.ShouldResemble, SexpBool{Val: true})
				cv.So(err, cv.ShouldEqual, nil)
			}

			// all sandSafeFuncs should be fine
			for name := range sandSafeFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? '%s)", name))
				cv.So(res, cv.ShouldResemble, SexpBool{Val: true})
				cv.So(err, cv.ShouldEqual, nil)
			}

		}
	})
}
