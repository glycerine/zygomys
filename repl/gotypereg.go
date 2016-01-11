package zygo

// The Go Type Registry
// ====================
//
// simply decide upon a name, and add a constructor
// function for that returns a pointer to your struct.
// The simply add to the init() function below.
//
// The env parameter to your MakeGoStructFunc()
// funcion is there is case you want to initialize
// your struct differently depending on the content
// of its context, but this is not common.

type MakeGoStructFunc func(env *Glisp) interface{}

var GostructRegistry = map[string]MakeGoStructFunc{}

// builtin known Go Structs
func init() {
	GostructRegistry["event"] = func(env *Glisp) interface{} {
		v := make([]Event, 1)
		return &v[0]
	}
	GostructRegistry["person"] = func(env *Glisp) interface{} {
		v := make([]Person, 1)
		return &v[0]
	}

	GostructRegistry["snoopy"] = func(env *Glisp) interface{} {
		v := make([]Snoopy, 1)
		return &v[0]
	}
	GostructRegistry["hornet"] = func(env *Glisp) interface{} {
		v := make([]Hornet, 1)
		return &v[0]
	}
	GostructRegistry["hellcat"] = func(env *Glisp) interface{} {
		v := make([]Hellcat, 1)
		return &v[0]
	}
}
