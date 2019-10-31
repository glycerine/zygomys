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

func BenchmarkCallUserFunction(b *testing.B) {
	env := NewZlisp()
	env.AddFunction("dosomething", func(*Zlisp, string, []Sexp) (r Sexp, err error) { return })
	script := fmt.Sprintf(`
		(for [(def i 0) (< i 1000000) (set i (+ i 1))]
			(dosomething)
		)
	`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.EvalString(script)
	}
}
