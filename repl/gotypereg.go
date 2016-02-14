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
var GoStructRegistry GoStructRegistryType

// the registry type
type GoStructRegistryType struct {
	Registry map[string]*RegistryEntry
}

// consistently ordered list of all registered types (created at init time).
var InitTimeListRegisteredTypes = []string{}

func (r *GoStructRegistryType) Register(name string, e *RegistryEntry) {
	InitTimeListRegisteredTypes = append(InitTimeListRegisteredTypes, name)
	r.Registry[name] = e
}

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
	GoStructRegistry = GoStructRegistryType{
		Registry: make(map[string]*RegistryEntry),
	}

	gsr := &GoStructRegistry

	// add go builtin types
	// ====================

	gsr.Register("byte", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(byte)
	}})
	gsr.Register("uint8", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint8)
	}})
	gsr.Register("uint16", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint16)
	}})
	gsr.Register("uint32", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint32)
	}})
	gsr.Register("uint64", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint64)
	}})
	gsr.Register("int8", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int8)
	}})
	gsr.Register("int16", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int16)
	}})
	gsr.Register("int32", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int32)
	}})
	gsr.Register("int64", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int64)
	}})
	gsr.Register("float32", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float32)
	}})

	gsr.Register("float64", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float64)
	}})

	gsr.Register("complex64", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex64)
	}})

	gsr.Register("complex128", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex128)
	}})

	gsr.Register("bool", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(bool)
	}})

	gsr.Register("string", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(string)
	}})

	gsr.Register("map[interface{}]interface{}", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[interface{}]interface{})
		return &m
	}})

	gsr.Register("map[string]interface{}", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[string]interface{})
		return &m
	}})

	gsr.Register("[]interface{}", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]interface{}, 0)
		return &m
	}})

	gsr.Register("[]string", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]string, 0)
		return &m
	}})

	gsr.Register("[]int64", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]int64, 0)
		return &m
	}})

	gsr.Register("time.Time", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(time.Time)
	}})

	gsr.Register("tm.Frame", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(tm.Frame)
	}})

	gsr.Register("[]tm.Frame", &RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]tm.Frame, 0)
		return &m
	}})

	// demo and user defined structs
	gsr.Register("event-demo", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Event{}
	}})
	gsr.Register("person-demo", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Person{}
	}})
	gsr.Register("snoopy", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Snoopy{}
	}})
	gsr.Register("hornet", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hornet{}
	}})
	gsr.Register("hellcat", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hellcat{}
	}})
	gsr.Register("weather", &RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Weather{}
	}})

	// cache all empty values and types
	for _, e := range gsr.Registry {
		e.ValueCache = reflect.ValueOf(e.Factory(nil))
		e.TypeCache = e.ValueCache.Type()
	}
}

func TypeListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 0 {
		return SexpNull, WrongNargs
	}
	r := InitTimeListRegisteredTypes
	s := make([]Sexp, len(r))
	for i := range r {
		s[i] = SexpStr(r[i])
	}
	return SexpArray(s), nil
}
