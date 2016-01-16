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
}

// builtin known Go Structs
// NB these are used to test the functionality of the
//    Go integration.
//
func init() {
	GostructRegistry["event-demo"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Event{}
	}}
	GostructRegistry["person-demo"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Person{}
	}}
	GostructRegistry["snoopy"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Snoopy{}
	}}
	GostructRegistry["hornet"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Hornet{}
	}}
	GostructRegistry["hellcat"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Hellcat{}
	}}
	GostructRegistry["weather"] = RegistryEntry{Factory: func(env *Glisp) interface{} {
		return &Weather{}
	}}
}
