package zygo

import (
	//"fmt"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test007ParentChildRecordsTranslateToGo(t *testing.T) {

	cv.Convey(`Given a tree of three records (hashes in zygomys); a snoopy`+
		` containing a hellcat, then SexpToGoStructs() should translate`+
		` that parent-child relationship faithfully into a Go Snoopy{}`+
		` containing a Go Hellcat{}.`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def snoop (snoopy chld:he))
`)
		panicOn(err)

		cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy chld: (hellcat speed:567))`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)
		cv.So(sn.Chld, cv.ShouldResemble, &Hellcat{Plane: Plane{Speed: 567}})
	})
}

func Test008CallByReflectionWorksWithoutNesting(t *testing.T) {

	cv.Convey(`Given an un-nested record without references to`+
		` other records; we should translate from record to Go`+
		` struct correctly`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def ho (hornet speed:567 nickname:"Bob" mass:4.2 SpanCm:8877))
`)
		panicOn(err)

		cv.So(x.SexpString(nil), cv.ShouldEqual,
			` (hornet speed:567 nickname:"Bob" mass:4.2 SpanCm:8877)`)

		ho := &Hornet{}
		res, err := SexpToGoStructs(x, ho, env)
		panicOn(err)
		VPrintf("\n ho = %#v\n", ho)
		VPrintf("\n res = %#v\n", res)
		cv.So(ho, cv.ShouldResemble, &Hornet{Plane: Plane{
			Wings: Wings{SpanCm: 8877}, Speed: 567},
			Nickname: "Bob", Mass: 4.2})
	})
}

func Test009CallByReflectionWorksWithoutNestingWithoutEmbeds(t *testing.T) {

	cv.Convey(`Given an un-nested record without references to other`+
		` records; and without embedded structs; we should translate`+
		` from record to Go struct correctly`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def ho (hornet nickname:"Bob" mass:4.2))
`)
		panicOn(err)

		cv.So(x.SexpString(nil), cv.ShouldEqual,
			` (hornet nickname:"Bob" mass:4.2)`)

		ho := &Hornet{}
		res, err := SexpToGoStructs(x, ho, env)
		panicOn(err)
		VPrintf("\n ho = %#v\n", ho)
		VPrintf("\n res = %#v\n", res)
		cv.So(ho, cv.ShouldResemble, &Hornet{Nickname: "Bob", Mass: 4.2})
	})
}

func Test010WriteIntoSingleInterfaceValueWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an interface scalar`+
		` value, this should translate from Sexp to Go correctly.`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet))
(def snoop (snoopy chld:he))
`)
		panicOn(err)

		cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy chld: (hellcat speed:567))`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)
		cv.So(sn.Chld, cv.ShouldResemble, &Hellcat{Plane: Plane{Speed: 567}})

	})
}

func Test011TranslationOfArraysWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of`+
		` concrete types, these should be translated from Sexp`+
		` correctly.`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def snoop (snoopy pack:[8 9 4]))
`)
		panicOn(err)

		cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy pack:[8 9 4])`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)
		cv.So(&sn, cv.ShouldResemble, &Snoopy{Pack: []int{8, 9, 4}})
	})
}

func Test012TranslationOfArraysOfInterfacesWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of Flyer`+
		` interfaces, these should be translated from Sexp correctly.`,
		t, func() {

			env := NewGlisp()
			defer env.parser.Stop()

			env.StandardSetup()

			x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy carrying:[he ho]))
`)
			panicOn(err)
			cv.So(x.SexpString(nil), cv.ShouldEqual,
				` (snoopy carrying:[ (hellcat speed:567)  (hornet SpanCm:12)])`)

			var sn Snoopy
			_, err = SexpToGoStructs(x, &sn, env)
			panicOn(err)
			VPrintf("\n sn = %#v\n", sn)
			cv.So(&sn, cv.ShouldResemble, &Snoopy{
				Carrying: []Flyer{
					&Hellcat{
						Plane: Plane{Speed: 567}},
					&Hornet{
						Plane: Plane{
							Wings: Wings{
								SpanCm: 12,
							},
						},
					}}})
		})
}

func Test014TranslationOfArraysOfInterfacesEmbeddedWorks(t *testing.T) {

	cv.Convey(`Given a parent Snoopy struct that has an array of Flyer`+
		` interfaces that are embedded inside Plane, these should be`+
		` translated from Sexp correctly.`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy friends:[he ho]))
`)
		panicOn(err)
		cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy friends:`+
			`[ (hellcat speed:567)  (hornet SpanCm:12)])`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)
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

func Test016ReflectCallOnGoMethodsZeroArgs(t *testing.T) {

	cv.Convey(`Given a translated to Go structs, we should be able`+
		` to invoke methods (zero in arguments) on them`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy friends:[he ho] cry:"yowza"))
`)
		panicOn(err)
		cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy friends:`+
			`[ (hellcat speed:567)  (hornet SpanCm:12)] cry:"yowza")`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)

		invok, err := env.EvalString(`
		   (_method snoop GetCry: )
		   `)
		panicOn(err)
		VPrintf("got invoke =  %T/val=%#v\n", invok, invok)

		switch arr := invok.(type) {
		case *SexpArray:
			// arr[0] should be string
			cv.So(arr.Val[0].(*SexpStr).S, cv.ShouldEqual, "yowza")
		default:
			VPrintf("got %T/val=%#v\n", arr, arr)
			panic("expected array back from _method")
		}

	})
}

func Test017ReflectCallOnGoMethodsOneArg(t *testing.T) {

	cv.Convey(`Given a translated to Go structs, we should be able to`+
		` invoke methods (with >= 1 in argument) on them`, t, func() {

		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy friends:[he ho] cry:"yowza"))
`)
		panicOn(err)
		cv.So(x.SexpString(nil), cv.ShouldEqual,
			` (snoopy friends:[ (hellcat speed:567)`+
				`  (hornet SpanCm:12)] cry:"yowza")`)

		var sn Snoopy
		_, err = SexpToGoStructs(x, &sn, env)
		panicOn(err)
		VPrintf("\n sn = %#v\n", sn)

		invok, err := env.EvalString(`
			   (_method snoop Fly: (weather time:(now) size:12 ` +
			`type:"sunny" details:(raw "123")))
			   `)
		panicOn(err)
		VPrintf("got invoke =  %T/val=%#v\n", invok, invok)

		switch arr := invok.(type) {
		case *SexpArray:
			// arr[0] should be string
			cv.So(arr.Val[0].(*SexpStr).S, cv.ShouldEqual,
				`Snoopy sees weather 'VERY sunny', cries 'yowza'`)
		default:
			VPrintf("got %T/val=%#v\n", arr, arr)
			panic("expected array back from _method")
		}

	})
}

func Test018ReflectCallOnGoMethodsComplexReturnType(t *testing.T) {

	cv.Convey(`Given a translated to Go structs, we should be able`+
		` to invoke methods on them that return struct pointers`,
		t, func() {

			env := NewGlisp()
			defer env.parser.Stop()

			env.StandardSetup()

			x, err := env.EvalString(`
(def he (hellcat speed:567))
(def ho (hornet SpanCm:12))
(def snoop (snoopy friends:[he ho] cry:"yowza"))
`)
			panicOn(err)
			cv.So(x.SexpString(nil), cv.ShouldEqual, ` (snoopy friends:`+
				`[ (hellcat speed:567)  (hornet SpanCm:12)] cry:"yowza")`)

			var sn Snoopy
			_, err = SexpToGoStructs(x, &sn, env)
			panicOn(err)
			VPrintf("\n sn = %#v\n", sn)

			invok, err := env.EvalString(`
			   (_method snoop EchoWeather: (weather time:(now) size:12 ` +
				`type:"sunny" details:(raw "123")))
			   `)
			panicOn(err)
			VPrintf("got invoke = '%s'\n", invok.SexpString(nil))
			cv.So(invok.SexpString(nil), cv.ShouldEqual, `[ (weather time:nil`+
				` size:12 type:"sunny" details:[]byte{0x31, 0x32, 0x33})]`)
		})
}
