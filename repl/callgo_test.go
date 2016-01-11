package zygo

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	//"github.com/shurcooL/go-goon"
	//"reflect"
	"testing"
)

func Test007CallByReflectionWorks(t *testing.T) {

	cv.Convey(`Given a tree of three records; a Snoopy{} containing a Hellcat{} and and Hornet{}: when the records point to each other inside an array, the shadow Go structs should also end up pointing at each other to form a mirror tree`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet))
(def snoop (snoopy chld:he))
`)
		panicOn(err)

		//cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy speed:105 chld:[ (hellcat speed:567)  (hornet )] cry:"yeeehaw")`)
		cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy chld: (hellcat speed:567))`)

		//_, err = GoLinkFunction(env, "golink", []Sexp{x})
		//panicOn(err)
		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		fmt.Printf("\n sn = %#v\n", sn)
		cv.So(sn.Chld, cv.ShouldResemble, &Hellcat{Plane: Plane{Speed: 567}})

		//		chld := (*(x.(SexpHash).GoStruct)).(*Snoopy).Chld
		//		cv.So(chld, cv.ShouldResemble, &Hellcat{})

		//x, err := env.EvalString(``)

		/*
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
		*/
	})
}

func Test008CallByReflectionWorksWithoutNesting(t *testing.T) {

	cv.Convey(`Given an un-nested record without references to other records; we should translate from record to Go struct correctly`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def ho (hornet speed:567 nickname:"Bob" mass:4.2 SpanCm:8877))
`)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (hornet speed:567 nickname:"Bob" mass:4.2 SpanCm:8877)`)

		ho := &Hornet{}
		res, err := SexpToGoStructs(x, ho, env)
		panicOn(err)
		fmt.Printf("\n ho = %#v\n", ho)
		fmt.Printf("\n res = %#v\n", res)
		cv.So(ho, cv.ShouldResemble, &Hornet{Plane: Plane{Wings: Wings{SpanCm: 8877}, Speed: 567}, Nickname: "Bob", Mass: 4.2})
	})
}

func Test009CallByReflectionWorksWithoutNestingWithoutEmbeds(t *testing.T) {

	cv.Convey(`Given an un-nested record without references to other records; and without embedded structs; we should translate from record to Go struct correctly`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def ho (hornet nickname:"Bob" mass:4.2))
`)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (hornet nickname:"Bob" mass:4.2)`)

		ho := &Hornet{}
		res, err := SexpToGoStructs(x, ho, env)
		panicOn(err)
		fmt.Printf("\n ho = %#v\n", ho)
		fmt.Printf("\n res = %#v\n", res)
		cv.So(ho, cv.ShouldResemble, &Hornet{Nickname: "Bob", Mass: 4.2})
	})
}

func Test010WriteIntoSingleInterfaceValueWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an interface scalar value, this should translate from Sexp to Go correctly.`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet))
(def snoop (snoopy chld:he))
`)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy chld: (hellcat speed:567))`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		fmt.Printf("\n sn = %#v\n", sn)
		cv.So(sn.Chld, cv.ShouldResemble, &Hellcat{Plane: Plane{Speed: 567}})

	})
}

func Test011TranslationOfArraysWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of concrete types, these should be translated from Sexp correctly.`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def snoop (snoopy pack:[8 9 4]))
`)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy pack:[8 9 4])`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		fmt.Printf("\n sn = %#v\n", sn)
		cv.So(&sn, cv.ShouldResemble, &Snoopy{Pack: []int{8, 9, 4}})
	})
}

func Test012TranslationOfArraysOfInterfacesWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of Flyer interfaces, these should be translated from Sexp correctly.`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy carrying:[he ho]))
`)
		panicOn(err)
		cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy carrying:[ (hellcat speed:567)  (hornet SpanCm:12)])`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		fmt.Printf("\n sn = %#v\n", sn)
		cv.So(&sn, cv.ShouldResemble, &Snoopy{Carrying: []Flyer{&Hellcat{Plane: Plane{Speed: 567}}, &Hornet{
			Plane: Plane{
				Wings: Wings{
					SpanCm: 12,
				},
			},
		}}})

	})
}

func Test014TranslationOfArraysOfInterfacesEmbeddedWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of Flyer interfaces that are embedded inside Plane, these should be translated from Sexp correctly.`, t, func() {

		env := NewGlisp()
		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy friends:[he ho]))
`)
		panicOn(err)
		cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy friends:[ (hellcat speed:567)  (hornet SpanCm:12)])`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		fmt.Printf("\n sn = %#v\n", sn)
		cv.So(&sn, cv.ShouldResemble, &Snoopy{
			Plane: Plane{
				Friends: []Flyer{
					&Hellcat{Plane: Plane{Speed: 567}},
					&Hornet{Plane: Plane{
						Wings: Wings{
							SpanCm: 12,
						},
					},
					},
				},
			},
		})

	})
}
