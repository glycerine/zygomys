package mytypes

import (
	"fmt"
	"os"
	"testing"

	"github.com/glycerine/glisp/interpreter"
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
func Test001TypeReflectionOnMytypeEvent(t *testing.T) {

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
		env := glisp.NewGlisp()
		env.StandardSetup()

		_, err := env.EvalString(activate)
		panicOn(err)

		_, err = env.EvalString(activate2)
		panicOn(err)

		x, err := env.EvalString(event)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (event id:123 user: (person first:"Liz" last:"C") flight:"AZD234" pilot:["Roger" "Ernie"])`)

		json := glisp.ToJson(x)
		cv.So(string(json), cv.ShouldEqual, `{"id":123, "user":{"first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)
		msgpack, goObj := glisp.ToMsgpack(x)
		expectedMsgpack := []byte{0x84, 0xa6, 0x66, 0x6c, 0x69, 0x67, 0x68, 0x74, 0xa6, 0x41, 0x5a, 0x44, 0x32, 0x33, 0x34, 0xa2, 0x69, 0x64, 0x7b, 0xa5, 0x70, 0x69, 0x6c, 0x6f, 0x74, 0x92, 0xa5, 0x52, 0x6f, 0x67, 0x65, 0x72, 0xa5, 0x45, 0x72, 0x6e, 0x69, 0x65, 0xa4, 0x75, 0x73, 0x65, 0x72, 0x82, 0xa5, 0x66, 0x69, 0x72, 0x73, 0x74, 0xa3, 0x4c, 0x69, 0x7a, 0xa4, 0x6c, 0x61, 0x73, 0x74, 0xa1, 0x43}
		cv.So(msgpack, cv.ShouldResemble, expectedMsgpack)

		_, goObj2 := glisp.MsgpackToJson(msgpack)
		// the ordering of jsonBack is canonical, so won't match ours
		// cv.So(string(jsonBack), cv.ShouldResemble, `{"id":123, "user":{"first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)

		fmt.Printf("goObj = '%#v'\n", goObj)
		fmt.Printf("goObj2 = '%#v'\n", goObj2)

		cv.So(goObj, cv.ShouldResemble, goObj2)

		f, err := os.Create("./test.msgpack")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		_, err = f.Write(msgpack)
		if err != nil {
			panic(err)
		}

	})
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
