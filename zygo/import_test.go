package zygo

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

func Test050ImportWorks(t *testing.T) {

	cv.Convey(`import test for https://github.com/glycerine/zygomys/issues/64`, t, func() {

		str := `(import  "../tests/foo.pkg")
                (assert (== foo.B "I am a Public string"))`

		dothings := func() error {
			var err error
			env := NewZlisp()
			env.StandardSetup()

			if err = env.LoadString(str); err != nil {
				panic(err)
			}
			expr, err := env.Run()
			_ = expr
			panicOn(err)
			//vv("expr = '%v'", expr.SexpString(nil))
			return nil
		}
		panicOn(dothings())
	})
}
