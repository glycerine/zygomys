package glisp

import (
	"fmt"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

/*
 Go map[string]interface{}  <--(1)--> lisp
   ^
   |
  (2)
   |
   V
 msgpack <--(3)--> go struct, strongly typed

(1) we test for here
(2) provided by ugorji/go/codec
(3) provided by tinylib/msgp, and by ugorji/go/codec

*/
func Test005ConversionToAndFromMsgpackAndJson(t *testing.T) {

	cv.Convey(`from gl we should be able to create a known Go struct,

type Event struct {
	Id     string
	User   string
	Flight string
	Pilot  []string
}

 Event{}, and fill in its fields`, t, func() {
		activate := `(defmap event)`
		activate2 := `(defmap person)`
		event := `(event id:123 user: (person first:"Liz" last:"C") flight:"AZD234"  pilot:["Roger" "Ernie"])`
		env := NewGlisp()
		env.StandardSetup()

		_, err := env.EvalString(activate)
		panicOn(err)

		_, err = env.EvalString(activate2)
		panicOn(err)

		x, err := env.EvalString(event)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (event id:123 user: (person first:"Liz" last:"C") flight:"AZD234" pilot:["Roger" "Ernie"])`)

		json := ToJson(x)
		//cv.So(string(json), cv.ShouldEqual, `{"Atype":"event", "id":123, "user":{"Atype":"person", "first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)
		cv.So(string(json), cv.ShouldEqual, `{"Atype":"event", "id":123, "user":{"Atype":"person", "first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)
		msgpack, goObj := ToMsgpack(x)
		//cv.So(msgpack, cv.ShouldResemble, expectedMsgpack)

		_, goObj2 := MsgpackToJson(msgpack)
		// the ordering of jsonBack is canonical, so won't match ours
		// cv.So(string(jsonBack), cv.ShouldResemble, `{"id":123, "user":{"first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)

		fmt.Printf("goObj = '%#v'\n", goObj)
		fmt.Printf("goObj2 = '%#v'\n", goObj2)

		cv.So(goObj, cv.ShouldResemble, goObj2)

		sexp, err := FromMsgpack(msgpack, env)
		panicOn(err)
		// must get into same order to have sane comparison, so borrow the KeyOrder to be sure.
		ko := sexp.(SexpHash).KeyOrder
		*ko = *x.(SexpHash).KeyOrder
		sexpStr := sexp.SexpString()
		cv.So(sexpStr, cv.ShouldResemble, ` (event id:123 user: (person first:"Liz" last:"C") flight:"AZD234" pilot:["Roger" "Ernie"])`)
	})
}
