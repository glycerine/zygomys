package mytypes

import (
	"testing"

	cv "github.com/glycerine/goconvey/convey"
	"github.com/zhemao/glisp/interpreter"
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
	Pilot  string
}

 Event{}, and fill in its fields`, t, func() {
		activate := `(msgpack-map event)`
		event := `(event id:"abc" user:"Liz" flight:"AZD234"  pilot:"Roger")`
		env := glisp.NewGlisp()

		_, err := env.EvalString(activate)
		panicOn(err)

		x, err := env.EvalString(event)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, `(event id:"abc" user:"Liz" flight:"AZD234"  pilot:"Roger")`)
	})
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
