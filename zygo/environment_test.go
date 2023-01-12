package zygo

import (
	"fmt"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test400SandboxFunctions(t *testing.T) {

	cv.Convey(`Given that the developer wishes to sandbox the Zygo interpreter when embedding it in their program, the NewZlispSandbox() function should return an interpreter that cannot call system/filesystem functions`, t, func() {

		sysFuncs := SystemFunctions()
		sandSafeFuncs := SandboxSafeFunctions()
		{
			env := NewZlispSandbox()

			// no system functions should pass
			for name := range sysFuncs {
				env.Clear()
				//P("checking name = '%v'", name)
				res, err := env.EvalString(fmt.Sprintf("(defined? %%%s)", name))
				cv.So(res, cv.ShouldResemble, &SexpBool{Val: false})
				cv.So(err, cv.ShouldResemble, nil)
			}

			// all sandSafeFuncs should be fine
			for name := range sandSafeFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? %%%s)", name))
				switch y := res.(type) {
				case *SexpSentinel:
					P("'%s' wasn't defined but should be; defined? returned '%s'", name, y.SexpString(nil))
				case *SexpBool:
					cv.So(res, cv.ShouldResemble, &SexpBool{Val: true})
				}
				cv.So(err, cv.ShouldEqual, nil)
			}
		}

		{
			fmt.Printf("\n and all functions should be reachable from a non-sandboxed environment.\n")
			env := NewZlisp()
			for name := range sysFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? %%%s)", name))
				cv.So(res, cv.ShouldResemble, &SexpBool{Val: true})
				cv.So(err, cv.ShouldEqual, nil)
			}

			// all sandSafeFuncs should be fine
			for name := range sandSafeFuncs {
				env.Clear()
				res, err := env.EvalString(fmt.Sprintf("(defined? %%%s)", name))
				cv.So(res, cv.ShouldResemble, &SexpBool{Val: true})
				cv.So(err, cv.ShouldEqual, nil)
			}

		}
	})
}

func TestCallUserFunction(t *testing.T) {
	cv.Convey(`It should recover from user-land panics and give stack traces`, t, func() {
		env := NewZlisp()
		env.AddFunction("dosomething", func(*Zlisp, string, []Sexp) (r Sexp, err error) {
			panic("I don't know how to do anything")
		})
		_, err := env.EvalString("(dosomething)")
		cv.So(err, cv.ShouldNotBeNil)
		cv.So(err.Error(), cv.ShouldContainSubstring, "stack trace:")
		cv.So(err.Error(), cv.ShouldContainSubstring, "github.com/glycerine/zygomys")
		cv.So(err.Error(), cv.ShouldContainSubstring, "zygo.(*Zlisp).CallUserFunction")
	})
}
