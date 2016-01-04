package glisp

import (
	"bytes"
	"reflect"

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
