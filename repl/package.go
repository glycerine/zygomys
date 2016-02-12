package zygo

import (
	"fmt"
)

// package.go: declare package, structs, function types

// A builder is a special kind of function. Like
// a macro it receives the un-evaluated tree
// of symbols from its caller. A builder
// can therefore be used to build new types
// and declarations new functions/methods.
//
// Like a function, a builder is called at
// run/evaluation time, not at definition time.
//
// The primary use here is to be able to define
// packges, structs, interfaces, functions,
// methods, and type aliases.
//
func (env *Glisp) ImportPackageBuilder() {
	env.AddBuilder("package", PackageBuilder)
}

// this is just a stub. TODO: finish design, implement packages.
func PackageBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("package name is missing. use: " +
			"(package package-name ...)\n")
	}

	P("in package builder, args = ")
	for i := range args {
		P("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}
