package glisp

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
)

func ToJson(exp Sexp) string {
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
	str := "{"
	for _, key := range *hash.KeyOrder {
		val, err := hash.HashGet(key)
		if err == nil {
			str += `"` + key.SexpString() + `":`
			str += string(ToJson(val)) + `, `
		} else {
			panic(err)
		}
	}
	if len(hash.Map) > 0 {
		return str[:len(str)-2] + "}"
	}
	return str + "}"
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
func ToMsgpack(exp Sexp) ([]byte, interface{}) {

	json := []byte(ToJson(exp))
	return JsonToMsgpack(json)
}

func JsonToMsgpack(json []byte) ([]byte, interface{}) {
	var iface interface{}

	decoder := codec.NewDecoderBytes(json, &msgpHelper.jh)
	err := decoder.Decode(&iface)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)

	var w bytes.Buffer
	enc := codec.NewEncoder(&w, &msgpHelper.mh)
	err = enc.Encode(&iface)
	if err != nil {
		panic(err)
	}

	return w.Bytes(), iface
}

func MsgpackToJson(msgp []byte) ([]byte, interface{}) {

	// msgpack -> go
	var iface interface{}
	dec := codec.NewDecoderBytes(msgp, &msgpHelper.mh)
	err := dec.Decode(&iface)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)

	// go -> json
	var w bytes.Buffer
	encoder := codec.NewEncoder(&w, &msgpHelper.jh)
	err = encoder.Encode(&iface)
	if err != nil {
		panic(err)
	}

	return w.Bytes(), iface
}

// returns both the msgpack []bytes and the go intermediary
func FromMsgpack(msgp []byte) Sexp {

	var iface interface{}
	dec := codec.NewDecoderBytes(msgp, &msgpHelper.mh)
	err := dec.Decode(&iface)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)

	return decodeGoToSexpHelper(iface, 0)
}

func decodeGoToSexpHelper(r interface{}, depth int) (s Sexp) {

	VPrintf("decodeHelper() at depth %d, decoded type is %T\n", depth, r)
	switch val := r.(type) {
	case string:
		VPrintf("depth %d found string case: val = %#v\n", depth, val)
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
			slice = append(slice, decodeGoToSexpHelper(val[i], depth+1))
		}
		return SexpArray(slice)

	case map[string]interface{}:

		VPrintf("depth %d found map[string]interface case: val = %#v\n", depth, val)
		sortedMapKey, sortedMapVal := makeSortedSlicesFromMap(val)

		pairs := make([]Sexp, 0)

		for i := range sortedMapKey {
			pairs = append(pairs, SexpStr(sortedMapKey[i]))
			ele := decodeGoToSexpHelper(sortedMapVal[i], depth+1)
			pairs = append(pairs, ele)
		}
		hash, err := MakeHash(pairs, "hash")
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
