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
var GostructRegistry = map[string]MakeGoStructFunc{}

// the type of all maker functions
type MakeGoStructFunc func(env *Glisp) interface{}

// builtin known Go Structs
// NB these are used to test the functionality of the
//    Go integration.
//
func init() {
	GostructRegistry["event-demo"] = func(env *Glisp) interface{} {
		return &Event{}
	}
	GostructRegistry["person-demo"] = func(env *Glisp) interface{} {
		return &Person{}
	}
	GostructRegistry["snoopy"] = func(env *Glisp) interface{} {
		return &Snoopy{}
	}
	GostructRegistry["hornet"] = func(env *Glisp) interface{} {
		return &Hornet{}
	}
	GostructRegistry["hellcat"] = func(env *Glisp) interface{} {
		return &Hellcat{}
	}
}
