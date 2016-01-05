package gdsl

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
)

/*
 Conversion map

 Go map[string]interface{}  <--(1)--> lisp
   ^                                  ^
   |                                 /
  (2)   ------------ (4) -----------/
   |   /
   V  V
 msgpack <--(3)--> go struct, strongly typed

(1) we provide these herein
     (a) SexpToGo()
     (b) GoToSexp()
(2) provided by ugorji/go/codec; see examples also herein
     (a) MsgpackToGo() / JsonToGo()
     (b) GoToMsgpack() / GoToJson()
(3) provided by tinylib/msgp, and by ugorji/go/codec
     by using pre-compiled or just decoding into an instance
     of the struct.
(4) see herein
     (a) SexpToMsgpack() and SexpToJson()
     (b) MsgpackToSexp(); uses (4) = (2) + (1)
*/
func JsonFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch name {
	case "json":
		str := SexpToJson(args[0])
		return SexpRaw([]byte(str)), nil
	case "unjson":
		raw, isRaw := args[0].(SexpRaw)
		if !isRaw {
			return SexpNull, fmt.Errorf("unjson error: SexpRaw required, but we got %T instead.", args[0])
		}
		return JsonToSexp([]byte(raw), env)
	case "msgpack":
		by, _ := SexpToMsgpack(args[0])
		return SexpRaw([]byte(by)), nil
	case "unmsgpack":
		raw, isRaw := args[0].(SexpRaw)
		if !isRaw {
			return SexpNull, fmt.Errorf("unmsgpack error: SexpRaw required, but we got %T instead.", args[0])
		}
		return MsgpackToSexp([]byte(raw), env)
	default:
		return SexpNull, fmt.Errorf("JsonFunction error: unrecognized function name: '%s'", name)
	}

	return nil, nil
}

// json -> sexp. env is needed to handle symbols correctly
func JsonToSexp(json []byte, env *Glisp) (Sexp, error) {
	iface, err := JsonToGo(json)
	if err != nil {
		return nil, err
	}
	return GoToSexp(iface, env)
}

// sexp -> json
func SexpToJson(exp Sexp) string {
	switch e := exp.(type) {
	case SexpHash:
		return e.jsonHashHelper()
	case SexpArray:
		return e.jsonArrayHelper()
	default:
		return exp.SexpString()
	}
}

func (hash *SexpHash) jsonHashHelper() string {
	str := fmt.Sprintf(`{"Atype":"%s", `, *hash.TypeName)

	ko := []string{}
	n := len(*hash.KeyOrder)
	if n == 0 {
		return str[:len(str)-2] + "}"
	}

	for _, key := range *hash.KeyOrder {
		keyst := key.SexpString()
		ko = append(ko, keyst)
		val, err := hash.HashGet(key)
		if err == nil {
			str += `"` + keyst + `":`
			str += string(SexpToJson(val)) + `, `
		} else {
			panic(err)
		}
	}

	str += `"zKeyOrder":[`
	for _, key := range ko {
		str += `"` + key + `", `
	}
	if n > 0 {
		str = str[:len(str)-2]
	}
	str += "]}"

	VPrintf("\n\n final ToJson() str = '%s'\n", str)
	return str
}

func (arr *SexpArray) jsonArrayHelper() string {
	if len(*arr) == 0 {
		return "[]"
	}

	str := "[" + (*arr)[0].SexpString()
	for _, sexp := range (*arr)[1:] {
		str += ", " + sexp.SexpString()
	}
	return str + "]"
}

type msgpackHelper struct {
	initialized bool
	mh          codec.MsgpackHandle
	jh          codec.JsonHandle
}

func (m *msgpackHelper) init() {
	if m.initialized {
		return
	}

	m.mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	// configure extensions
	// e.g. for msgpack, define functions and enable Time support for tag 1
	//does this make a differenece? m.mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	m.mh.RawToString = true
	m.mh.WriteExt = true
	m.mh.SignedInteger = true
	m.mh.Canonical = true // sort maps before writing them

	// JSON
	m.jh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	m.jh.SignedInteger = true
	m.jh.Canonical = true // sort maps before writing them

	m.initialized = true
}

var msgpHelper msgpackHelper

func init() {
	msgpHelper.init()
}

// translate to sexp -> json -> go -> msgpack
// returns both the msgpack []bytes and the go intermediary
func SexpToMsgpack(exp Sexp) ([]byte, interface{}) {

	json := []byte(SexpToJson(exp))
	iface, err := JsonToGo(json)
	panicOn(err)
	by, err := GoToMsgpack(iface)
	panicOn(err)
	return by, iface
}

// json -> go
func JsonToGo(json []byte) (interface{}, error) {
	var iface interface{}

	decoder := codec.NewDecoderBytes(json, &msgpHelper.jh)
	err := decoder.Decode(&iface)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)

	return iface, nil
}

func GoToMsgpack(iface interface{}) ([]byte, error) {
	var w bytes.Buffer
	enc := codec.NewEncoder(&w, &msgpHelper.mh)
	err := enc.Encode(&iface)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// go -> json
func GoToJson(iface interface{}) []byte {
	var w bytes.Buffer
	encoder := codec.NewEncoder(&w, &msgpHelper.jh)
	err := encoder.Encode(&iface)
	if err != nil {
		panic(err)
	}
	return w.Bytes()
}

// msgpack -> sexp
func MsgpackToSexp(msgp []byte, env *Glisp) (Sexp, error) {
	iface, err := MsgpackToGo(msgp)
	if err != nil {
		return nil, fmt.Errorf("MsgpackToSexp failed at MsgpackToGo step: '%s", err)
	}
	sexp, err := GoToSexp(iface, env)
	if err != nil {
		return nil, fmt.Errorf("MsgpackToSexp failed at GoToSexp step: '%s", err)
	}
	return sexp, nil
}

// msgpack -> go
func MsgpackToGo(msgp []byte) (interface{}, error) {

	var iface interface{}
	dec := codec.NewDecoderBytes(msgp, &msgpHelper.mh)
	err := dec.Decode(&iface)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)
	return iface, nil
}

// convert iface, which will typically be map[string]interface{},
// into an s-expression
func GoToSexp(iface interface{}, env *Glisp) (Sexp, error) {
	return decodeGoToSexpHelper(iface, 0, env, false), nil
}

func decodeGoToSexpHelper(r interface{}, depth int, env *Glisp, preferSym bool) (s Sexp) {

	VPrintf("decodeHelper() at depth %d, decoded type is %T\n", depth, r)
	switch val := r.(type) {
	case string:
		VPrintf("depth %d found string case: val = %#v\n", depth, val)
		if preferSym {
			return env.MakeSymbol(val)
		}
		return SexpStr(val)

	case int:
		VPrintf("depth %d found int case: val = %#v\n", depth, val)
		return SexpInt(val)

	case int32:
		VPrintf("depth %d found int32 case: val = %#v\n", depth, val)
		return SexpInt(val)

	case int64:
		VPrintf("depth %d found int64 case: val = %#v\n", depth, val)
		return SexpInt(val)

	case float64:
		VPrintf("depth %d found float64 case: val = %#v\n", depth, val)
		return SexpFloat(val)

	case []interface{}:
		VPrintf("depth %d found []interface{} case: val = %#v\n", depth, val)

		slice := []Sexp{}
		for i := range val {
			slice = append(slice, decodeGoToSexpHelper(val[i], depth+1, env, preferSym))
		}
		return SexpArray(slice)

	case map[string]interface{}:

		VPrintf("depth %d found map[string]interface case: val = %#v\n", depth, val)
		sortedMapKey, sortedMapVal := makeSortedSlicesFromMap(val)

		pairs := make([]Sexp, 0)

		typeName := "hash"
		var keyOrd Sexp
		foundzKeyOrder := false
		for i := range sortedMapKey {
			// special field storing the name of our record (defmap) type.
			VPrintf("\n i=%d sortedMapVal type %T, value=%v\n", i, sortedMapVal[i], sortedMapVal[i])
			VPrintf("\n i=%d sortedMapKey type %T, value=%v\n", i, sortedMapKey[i], sortedMapKey[i])
			if sortedMapKey[i] == "zKeyOrder" {
				keyOrd = decodeGoToSexpHelper(sortedMapVal[i], depth+1, env, true)
				foundzKeyOrder = true
			} else if sortedMapKey[i] == "Atype" {
				tn, isString := sortedMapVal[i].(string)
				if isString {
					typeName = string(tn)
				}
			} else {
				sym := env.MakeSymbol(sortedMapKey[i])
				pairs = append(pairs, sym)
				ele := decodeGoToSexpHelper(sortedMapVal[i], depth+1, env, preferSym)
				pairs = append(pairs, ele)
			}
		}
		hash, err := MakeHash(pairs, typeName)
		if foundzKeyOrder {
			err = SetHashKeyOrder(&hash, keyOrd)
			panicOn(err)
		}
		panicOn(err)
		return hash

	case []byte:
		VPrintf("depth %d found []byte case: val = %#v\n", depth, val)

		return SexpRaw(val)

	case nil:
		return SexpNull

	case bool:
		return SexpBool(val)

	default:
		fmt.Printf("unknown type in type switch, val = %#v.  type = %T.\n", val, val)
	}

	return s
}

//msgp:ignore mapsorter KiSlice

type mapsorter struct {
	key   string
	iface interface{}
}

type KiSlice []*mapsorter

func (a KiSlice) Len() int           { return len(a) }
func (a KiSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a KiSlice) Less(i, j int) bool { return a[i].key < a[j].key }

func makeSortedSlicesFromMap(m map[string]interface{}) ([]string, []interface{}) {
	key := make([]string, len(m))
	val := make([]interface{}, len(m))
	so := make(KiSlice, 0)
	for k, i := range m {
		so = append(so, &mapsorter{key: k, iface: i})
	}
	sort.Sort(so)
	for i := range so {
		key[i] = so[i].key
		val[i] = so[i].iface
	}
	return key, val
}

// translate an Sexpr to a go value that doesn't
// depend on any Sexp/Glisp types. Glisp maps
// will get turned into map[string]interface{}.
// This is mostly just an exercise in type conversion.
func SexpToGo(sexp Sexp, env *Glisp) interface{} {

	switch e := sexp.(type) {
	case SexpRaw:
		return []byte(e)
	case SexpArray:
		ar := make([]interface{}, len(e))
		for i, ele := range e {
			ar[i] = SexpToGo(ele, env)
		}
		return ar
	case SexpInt:
		// ugorji msgpack will give us int64 not int,
		// so match that to make the decodings comparable.
		return int64(e)
	case SexpStr:
		return string(e)
	case SexpChar:
		return rune(e)
	case SexpFloat:
		return float64(e)
	case SexpHash:
		m := make(map[string]interface{})
		for _, arr := range e.Map {
			for _, pair := range arr {
				key := SexpToGo(pair.head, env)
				val := SexpToGo(pair.tail, env)
				keyString, isStringKey := key.(string)
				if !isStringKey {
					panic(fmt.Errorf("key '%v' should have been a string, but was not.", key))
				}
				m[keyString] = val
			}
		}
		m["Atype"] = *e.TypeName
		ko := make([]interface{}, 0)
		for _, k := range *e.KeyOrder {
			ko = append(ko, SexpToGo(k, env))
		}
		m["zKeyOrder"] = ko
		return m
	case SexpPair:
		// no conversion
		return e
	case SexpSymbol:
		return e.name
	case SexpFunction:
		// no conversion done
		return e
	case SexpSentinel:
		// no conversion done
		return e
	default:
		fmt.Printf("\n error: unknown type: %T in '%#v'\n", e, e)
	}
	return nil
}
