package zygo

import (
	tm "github.com/glycerine/tmframe"
	"reflect"
	"time"
)

// The Go Type Registry
// ====================
//
// simply decide upon a name, and add a maker
// function for that returns a pointer to your struct.
// The simply add to the init() function below.
//
// The env parameter to your MakeGoStructFunc()
// function is there is case you want to initialize
// your struct differently depending on the content
// of its context, but this is not commonly needed.
// Also, the factory method *must* support the
// env parameter being nil and still return a
// sensible, usable value. The factory will be called
// with env = nil during init() time.
//
// The repl will automatically do a (defmap record)
// for each record defined in the registry. e.g.
// for snoopy, hornet, hellcat, etc.
//
var GoStructRegistry = map[string]*RegistryEntry{}

// the type of all maker functions
type MakeGoStructFunc func(env *Glisp) interface{}

type RegistryEntry struct {
	Factory    MakeGoStructFunc
	Gen        bool // generate a defmap mapping?
	ValueCache reflect.Value
	TypeCache  reflect.Type
}

// builtin known Go Structs
// NB these are used to test the functionality of the
//    Go integration.
//
func init() {
	GoStructRegistry["event-demo"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Event{}
	}}
	GoStructRegistry["person-demo"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Person{}
	}}
	GoStructRegistry["snoopy"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Snoopy{}
	}}
	GoStructRegistry["hornet"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hornet{}
	}}
	GoStructRegistry["hellcat"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hellcat{}
	}}
	GoStructRegistry["weather"] = &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Weather{}
	}}

	// add go builtin types
	// ====================

	GoStructRegistry["byte"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(byte)
	}}
	GoStructRegistry["uint8"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint8)
	}}
	GoStructRegistry["uint16"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint16)
	}}
	GoStructRegistry["uint32"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint32)
	}}
	GoStructRegistry["uint64"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint64)
	}}
	GoStructRegistry["int8"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int8)
	}}
	GoStructRegistry["int16"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int16)
	}}
	GoStructRegistry["int32"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int32)
	}}
	GoStructRegistry["int64"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int64)
	}}
	GoStructRegistry["float32"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float32)
	}}

	GoStructRegistry["float64"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float64)
	}}

	GoStructRegistry["complex64"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex64)
	}}

	GoStructRegistry["complex128"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex128)
	}}

	GoStructRegistry["bool"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(bool)
	}}

	GoStructRegistry["string"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(string)
	}}

	GoStructRegistry["map[interface{}]interface{}"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[interface{}]interface{})
		return &m
	}}

	GoStructRegistry["map[string]interface{}"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[string]interface{})
		return &m
	}}

	GoStructRegistry["[]interface{}"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]interface{}, 0)
		return &m
	}}

	GoStructRegistry["[]string"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]string, 0)
		return &m
	}}

	GoStructRegistry["[]int64"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]int64, 0)
		return &m
	}}

	GoStructRegistry["time.Time"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(time.Time)
	}}

	GoStructRegistry["tm.Frame"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(tm.Frame)
	}}

	GoStructRegistry["[]tm.Frame"] = &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]tm.Frame, 0)
		return &m
	}}

	// cache all empty values and types
	for _, e := range GoStructRegistry {
		e.ValueCache = reflect.ValueOf(e.Factory(nil))
		e.TypeCache = e.ValueCache.Type()
	}
}

func ListRegisteredTypes() (res []string) {
	for k := range GoStructRegistry {
		res = append(res, k)
	}
	return
}

func TypeListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 0 {
		return SexpNull, WrongNargs
	}
	r := ListRegisteredTypes()
	s := make([]Sexp, len(r))
	for i := range r {
		s[i] = SexpStr(r[i])
	}
	return SexpArray(s), nil
}
