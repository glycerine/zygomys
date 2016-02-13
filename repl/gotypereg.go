package zygo

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
//
// The repl will automatically do a (defmap record)
// for each record defined in the registry. e.g.
// for snoopy, hornet, hellcat, etc.
//
var GostructRegistry = map[string]RegistryEntry{}

// the type of all maker functions
type MakeGoStructFunc func(env *Glisp) interface{}

type RegistryEntry struct {
	Factory MakeGoStructFunc
	Gen     bool // generate a defmap mapping?
}

// builtin known Go Structs
// NB these are used to test the functionality of the
//    Go integration.
//
func init() {
	GostructRegistry["event-demo"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Event{}
	}}
	GostructRegistry["person-demo"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Person{}
	}}
	GostructRegistry["snoopy"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Snoopy{}
	}}
	GostructRegistry["hornet"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hornet{}
	}}
	GostructRegistry["hellcat"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Hellcat{}
	}}
	GostructRegistry["weather"] = RegistryEntry{Gen: true, Factory: func(env *Glisp) interface{} {
		return &Weather{}
	}}

	// add go builtin types
	GostructRegistry["byte"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(byte)
	}}
	GostructRegistry["uint8"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint8)
	}}
	GostructRegistry["uint16"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint16)
	}}
	GostructRegistry["uint32"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint32)
	}}
	GostructRegistry["uint64"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(uint64)
	}}
	GostructRegistry["int8"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int8)
	}}
	GostructRegistry["int16"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int16)
	}}
	GostructRegistry["int32"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int32)
	}}
	GostructRegistry["int64"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(int64)
	}}
	GostructRegistry["float32"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float32)
	}}

	GostructRegistry["float64"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(float64)
	}}

	GostructRegistry["complex64"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex64)
	}}

	GostructRegistry["complex128"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(complex128)
	}}

	GostructRegistry["bool"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(bool)
	}}

	GostructRegistry["string"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		return new(string)
	}}

	GostructRegistry["map[interface{}]interface{}"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[interface{}]interface{})
		return &m
	}}

	GostructRegistry["map[string]interface{}"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make(map[string]interface{})
		return &m
	}}

	GostructRegistry["[]interface{}"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]interface{}, 0)
		return &m
	}}

	GostructRegistry["[]string"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]string, 0)
		return &m
	}}

	GostructRegistry["[]int64"] = RegistryEntry{Gen: false, Factory: func(env *Glisp) interface{} {
		m := make([]int64, 0)
		return &m
	}}

}
