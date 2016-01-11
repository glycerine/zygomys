package zygo

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"github.com/shurcooL/go-goon"
	"reflect"
	"testing"
)

func Test007CallByReflectionWorks(t *testing.T) {

	cv.Convey(`Given a Snoopy{} struct s, and a Weather{} struct w, use reflection to call s.Fly(w)`, t, func() {
		var s = &Snoopy{Cry: "ImaDog!"}
		var w = &Weather{Type: "thundery"}
		//method := s.Sideeffect
		//vmethod := reflect.ValueOf(method)
		//expected, err := s.Fly(w)

		//panicOn(err)
		//fmt.Printf("expected = '%#v'\n", expected)

		vw := reflect.ValueOf(w)
		in := []reflect.Value{vw}
		fmt.Printf("in = '%#v'\n", in)

		vs := reflect.ValueOf(s)
		fmt.Printf("vs = '%#v'\n", vs)

		// find the method on s
		ty := vs.Type()
		n := ty.NumMethod()
		fmt.Printf("n = %v\n", n) // 2 yes!
		m1 := ty.Method(0)
		fmt.Printf("m1.Name = %v\n", m1.Name) // Fly
		m2 := ty.Method(1)
		fmt.Printf("m2.Name = %v\n", m2.Name) // Sideeffect

		inObj1st := []reflect.Value{vs, vw}
		res := m1.Func.Call(inObj1st)
		//res := vmethod.Call(inempty)
		fmt.Printf("res = '%#v'\n", res)
		goon.Dump(res)
	})
}
