package zygo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
	"github.com/ugorji/go/codec"
)

/*
 we test for here (1)
     (a) SexpToGo()
     (b) GoToSexp()
*/
func Test005ConversionToAndFromMsgpackAndJson(t *testing.T) {

	cv.Convey(`from gl we should be able to create a known Go struct,

type Event struct {
	Id     int
	User   Person
	Flight string
	Pilot  []string
}

 Event{}, and fill in its fields`, t, func() {
		event := `(event-demo id:123 user: (person-demo first:"Liz" last:"C") flight:"AZD234"  pilot:["Roger" "Ernie"])`
		env := NewGlisp()
		defer env.parser.Stop()

		env.StandardSetup()

		x, err := env.EvalString(event)
		panicOn(err)

		cv.So(x.SexpString(), cv.ShouldEqual, ` (event-demo id:123 user: (person-demo first:"Liz" last:"C") flight:"AZD234" pilot:["Roger" "Ernie"])`)

		jsonBy := SexpToJson(x)
		cv.So(string(jsonBy), cv.ShouldEqual, `{"Atype":"event-demo", "id":123, "user":{"Atype":"person-demo", "first":"Liz", "last":"C", "zKeyOrder":["first", "last"]}, "flight":"AZD234", "pilot":["Roger", "Ernie"], "zKeyOrder":["id", "user", "flight", "pilot"]}`)
		msgpack, goObj := SexpToMsgpack(x)
		// msgpack field ordering is random, so can't expect a match the serialization byte-for-byte
		//cv.So(msgpack, cv.ShouldResemble, expectedMsgpack)

		goObj2, err := MsgpackToGo(msgpack)
		panicOn(err)
		// the ordering of jsonBack is canonical, so won't match ours
		// cv.So(string(jsonBack), cv.ShouldResemble, `{"id":123, "user":{"first":"Liz", "last":"C"}, "flight":"AZD234", "pilot":["Roger", "Ernie"]}`)

		fmt.Printf("goObj = '%#v'\n", goObj)
		fmt.Printf("goObj2 = '%#v'\n", goObj2)

		cv.So(goObj, cv.ShouldResemble, goObj2)

		iface, err := MsgpackToGo(msgpack)
		panicOn(err)
		sexp, err := GoToSexp(iface, env)
		panicOn(err)
		// must get into same order to have sane comparison, so borrow the KeyOrder to be sure.
		ko := sexp.(SexpHash).KeyOrder
		*ko = *x.(SexpHash).KeyOrder
		sexpStr := sexp.SexpString()
		expectedSexpr := ` (event-demo id:123 user: (person-demo first:"Liz" last:"C") flight:"AZD234" pilot:["Roger" "Ernie"])`
		cv.So(sexpStr, cv.ShouldResemble, expectedSexpr)

		fmt.Printf("\n Unmarshaling from msgpack into pre-defined go struct should succeed.\n")

		var goEvent Event
		dec := codec.NewDecoderBytes(msgpack, &msgpHelper.mh)
		err = dec.Decode(&goEvent)
		panicOn(err)
		fmt.Printf("from msgpack, goEvent = '%#v'\n", goEvent)
		cv.So(goEvent.Id, cv.ShouldEqual, 123)
		cv.So(goEvent.Flight, cv.ShouldEqual, "AZD234")
		cv.So(goEvent.Pilot[0], cv.ShouldEqual, "Roger")
		cv.So(goEvent.Pilot[1], cv.ShouldEqual, "Ernie")
		cv.So(goEvent.User.First, cv.ShouldEqual, "Liz")
		cv.So(goEvent.User.Last, cv.ShouldEqual, "C")

		goEvent = Event{}
		jdec := codec.NewDecoderBytes([]byte(jsonBy), &msgpHelper.jh)
		err = jdec.Decode(&goEvent)
		panicOn(err)
		fmt.Printf("from json, goEvent = '%#v'\n", goEvent)
		cv.So(goEvent.Id, cv.ShouldEqual, 123)
		cv.So(goEvent.Flight, cv.ShouldEqual, "AZD234")
		cv.So(goEvent.Pilot[0], cv.ShouldEqual, "Roger")
		cv.So(goEvent.Pilot[1], cv.ShouldEqual, "Ernie")
		cv.So(goEvent.User.First, cv.ShouldEqual, "Liz")
		cv.So(goEvent.User.Last, cv.ShouldEqual, "C")

		fmt.Printf("\n And directly from Go to S-expression via GoToSexp() should work.\n")
		sexp2, err := GoToSexp(goObj2, env)
		cv.So(sexp2.SexpString(), cv.ShouldEqual, expectedSexpr)
		fmt.Printf("\n Result: directly from Go map[string]interface{} -> sexpr via GoMapToSexp() produced: '%s'\n", sexp2.SexpString())

		fmt.Printf("\n And the reverse direction, from S-expression to go map[string]interface{} should work.\n")
		goMap3 := SexpToGo(sexp2, env).(map[string]interface{})

		// detailed diff
		goObj2map := goObj2.(map[string]interface{})

		// looks like goMap3 has an int, whereas goObj2map has an int64

		// compare goMap3 and goObj2
		for k3, v3 := range goMap3 {
			v2 := goObj2map[k3]
			cv.So(v3, cv.ShouldResemble, v2)
		}

		fmt.Printf("\n Directly Sexp -> msgpack -> pre-established Go struct Event{} should work.\n")

		switch asHash := sexp2.(type) {
		default:
			err = fmt.Errorf("value must be a hash or defmap")
			panic(err)
		case SexpHash:
			tn := *(asHash.TypeName)
			factory, hasMaker := GostructRegistry[tn]
			if !hasMaker {
				err = fmt.Errorf("type '%s' not registered in GostructRegistry", tn)
				panic(err)
			}
			newStruct := factory.Factory(env)

			// What didn't work here was going through msgpack, because
			// ugorji msgpack encode will write turn signed ints into unsigned ints,
			// which is a problem for msgp decoding. Hence cut out the middle men
			// and decode straight from jsonBytes into our newStruct.
			jsonBytes := []byte(SexpToJson(asHash))

			jsonDecoder := json.NewDecoder(bytes.NewBuffer(jsonBytes))
			err := jsonDecoder.Decode(newStruct)
			switch err {
			case io.EOF:
			case nil:
			default:
				panic(fmt.Errorf("error during jsonDecoder.Decode() on type '%s': '%s'", tn, err))
			}
			asHash.SetGoStructFactory(factory)

			fmt.Printf("from json via factory.Make(), newStruct = '%#v'\n", newStruct)
			cv.So(newStruct, cv.ShouldResemble, &goEvent)
		}
	})
}
