package zygo

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type SexpUserStructDefn struct {
	Name   string
	Fields []*SexpField
}

// pretty print a struct
func (p *SexpUserStructDefn) SexpString() string {
	if len(p.Fields) == 0 {
		return fmt.Sprintf("(struct %s)", p.Name)
	}
	s := fmt.Sprintf("(struct %s [\n", p.Name)

	w := make([][]int, len(p.Fields))
	maxfield := 0
	for j, f := range p.Fields {
		w[j] = f.FieldWidths()
		maxfield = maxi(maxfield, len(w[j]))
	}

	// computing padding
	// x
	// xx xx
	// xxxxxxx x
	// xxx x x x
	//
	// becomes
	//
	// x
	// xx      xx
	// xxxxxxx
	// xxx     x  x x
	pad := make([]int, maxfield)
	for j := range w {
		cur := 0
		for i := 0; i < maxfield; i++ {
			if i < len(w[j]) {
				cur = maxi(cur, w[j][i])
			}
		}
		pad = append(pad, cur)
	}
	for _, f := range p.Fields {
		s += "        " + f.AlignString(pad) + "\n"
	}
	s += "        ])\n"
	return s
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type SexpField SexpHash

// specialize for nice looking field prints
func (f *SexpField) FieldWidths() []int {
	hash := (*SexpHash)(f)
	wide := []int{}
	for _, key := range hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		str := ""
		if err == nil {
			switch s := key.(type) {
			case SexpStr:
				str += s.S + ":"
			case SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString() + ":"
			}
			str += val.SexpString() + " "
			wide = append(wide, len(str))
		} else {
			panic(err)
		}
	}
	return wide
}

func (f *SexpField) AlignString(pad []int) string {
	hash := (*SexpHash)(f)
	str := " (" + hash.TypeName + " "

	for i, key := range hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		if err == nil {
			switch s := key.(type) {
			case SexpStr:
				str += s.S + ":"
			case SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString() + ":"
			}
			if i > 0 {
				str += val.SexpString() + " "
			} else {
				str += val.SexpString() + strings.Repeat(" ", pad[i%len(pad)])
			}
		} else {
			panic(err)
		}
	}
	if len(hash.Map) > 0 {
		return str[:len(str)-1] + ")"
	}
	return str + ")"
}

func (f *SexpField) SexpString() string {
	hash := (*SexpHash)(f)
	str := " (" + hash.TypeName + " "

	for i, key := range hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		if err == nil {
			switch s := key.(type) {
			case SexpStr:
				str += s.S + ":"
			case SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString() + ":"
			}
			if i > 0 {
				str += val.SexpString() + " "
			} else {
				str += val.SexpString() + "    "
			}
		} else {
			panic(err)
		}
	}
	if len(hash.Map) > 0 {
		return str[:len(str)-1] + ")"
	}
	return str + ")"
}

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
	env.AddBuilder("struct", StructBuilder)
	env.AddBuilder("func", FuncBuilder)
	env.AddBuilder("interface", InterfaceBuilder)
	env.AddBuilder("package", PackageBuilder)

	env.AddFunction("slice-of", SliceOfFunction)
	env.AddFunction("pointer-to", PointerToFunction)
}

func StructBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	n := len(args)
	if n < 1 {
		return SexpNull, fmt.Errorf("struct name is missing. use: " +
			"(struct struct-name ...)\n")
	}

	P("in struct builder, name = '%s', args = ", name)
	for i := range args {
		P("args[%v] = '%s' of type %T", i, args[i].SexpString(), args[i])
	}
	var symN SexpSymbol
	switch b := args[0].(type) {
	case SexpSymbol:
		symN = b
	case SexpPair:
		sy, isQuo := isQuotedSymbol(b)
		if isQuo {
			symN = sy.(SexpSymbol)
		} else {
			return SexpNull, fmt.Errorf("bad struct name: symbol required")
		}
	default:
		return SexpNull, fmt.Errorf("bad struct name: symbol required")
	}

	/*
		sn, err := env.EvalExpressions(args[0:1])
		if err != nil {
			return SexpNull, fmt.Errorf("bad struct name: '%v'", err)
		}
		P("done with sn eval")
		symN, isSym := sn.(SexpSymbol)
		if !isSym {
			return SexpNull, fmt.Errorf("bad struct name: symbol required")
		}
	*/
	P("good: have struct name '%v'", symN)

	env.datastack.PushExpr(SexpNull)
	structName := symN.name

	var xar []Sexp
	var flat []*SexpField
	if n > 2 {
		return SexpNull, fmt.Errorf("bad struct declaration: more than two arguments." +
			"prototype is (struct name [(field ...)*] )")
	}
	if n == 2 {
		P("in case n == 2")
		switch ar := args[1].(type) {
		default:
			return SexpNull, fmt.Errorf("bad struct declaration '%v': second argument "+
				"must be a slice of fields."+
				" prototype is (struct name [(field ...)*] )", structName)
		case SexpArray:
			arr := []Sexp(ar)
			if len(arr) == 0 {
				// allow this
			} else {
				// dup to avoid messing with the stack on eval:
				dup := env.Duplicate()
				for i, ele := range arr {
					P("about to eval i=%v", i)
					ev, err := dup.EvalExpressions([]Sexp{ele})
					P("done with eval i=%v. ev=%v", i, ev.SexpString())
					if err != nil {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field at array entry %v; error was '%v'", structName, i, err)
					}
					P("checking for isHash at i=%v", i)
					asHash, isHash := ev.(*SexpField)
					if !isHash {
						P("was not hash, instead was %T", ev)
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; a (field ...) is required. Instead saw '%T'/with value = '%v'",
							structName, i, ev, ev.SexpString())
					}
					P("good eval i=%v, ev=%#v / %v", i, ev, ev.SexpString())
					ko := asHash.KeyOrder
					if len(ko) == 0 {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; field had no name",
							structName, i)
					}
					P("ko = '%#v'", ko)
					first := ko[0]
					P("first = '%#v'", first)
					xar = append(xar, first)
					xar = append(xar, ev)
					flat = append(flat, ev.(*SexpField))
				}
				P("no err from EvalExpressions, got xar = '%#v'", xar)
			}
		}
		/*
				P("evaluating args[1:2] which is of type %T / val=%#v", args[1], args[1])
				arr, err := env.EvalExpressions(args[1:2])
				if err != nil {
					return SexpNull, fmt.Errorf("bad struct declaration: bad "+
						"array of fields, error was '%v'", err)
				}

			switch ar := arr.(type) {
			case SexpArray:
				P("good, have array %#v", ar)
				xar = []Sexp(ar)
			default:
				return SexpNull, fmt.Errorf("bad struct declaration: did not find "+
					"array of fields following name; instead found %v/type=%T", ar, ar)
			}
		*/
	} // end n == 2

	uds := &SexpUserStructDefn{Name: structName, Fields: flat}
	P("good: made typeDefnHash: '%s'", uds.SexpString())
	rt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return uds, nil
	})
	rt.UserStructDefn = uds
	GoStructRegistry.RegisterUserdef(structName, rt, false)
	P("good: registered new userdefined struct '%s'", structName)
	err := env.LexicalBindSymbol(symN, rt)
	if err != nil {
		return SexpNull, fmt.Errorf("struct builder could not bind symbol '%': '%v'",
			structName, err)
	}
	P("good: bound symbol '%s' to RegisteredType '%s'", symN.SexpString(), rt.SexpString())
	return rt, nil
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

func InterfaceBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("interface name is missing. use: " +
			"(interface interface-name ...)\n")
	}

	P("in interface builder, args = ")
	for i := range args {
		P("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}

func FuncBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("func name is missing. use: " +
			"(func func-name ...)\n")
	}

	P("in func builder, args = ")
	for i := range args {
		P("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}

func SliceOfFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument to slice-of is missing. use: " +
			"(slice-of a-regtype)\n")
	}

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	default:
		return SexpNull, fmt.Errorf("argument to slice-of was not regtype, "+
			"instead type %T displaying as '%v' ",
			arg, arg.SexpString())
	}

	//P("slice-of arg = '%s' with type %T", args[0].SexpString(), args[0])

	derivedType := reflect.SliceOf(rt.TypeCache)
	sliceRt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return reflect.MakeSlice(derivedType, 0, 0), nil
	})
	sliceRt.DisplayAs = fmt.Sprintf("(slice-of %s)", rt.DisplayAs)
	sliceName := "slice-of-" + rt.RegisteredName
	GoStructRegistry.RegisterUserdef(sliceName, sliceRt, false)
	return sliceRt, nil
}

func PointerToFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument to pointer-to is missing. use: " +
			"(pointer-to a-regtype)\n")
	}

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	default:
		return SexpNull, fmt.Errorf("argument to pointer-to was not regtype, "+
			"instead type %T displaying as '%v' ",
			arg, arg.SexpString())
	}

	//P("pointer-to arg = '%s' with type %T", args[0].SexpString(), args[0])

	derivedType := reflect.PtrTo(rt.TypeCache)
	sliceRt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return reflect.New(derivedType), nil
	})
	sliceRt.DisplayAs = fmt.Sprintf("(pointer-to %s)", rt.DisplayAs)
	sliceName := "pointer-to-" + rt.RegisteredName
	GoStructRegistry.RegisterUserdef(sliceName, sliceRt, false)
	return sliceRt, nil
}

func StructConstructorFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	P("in struct ctor, name = '%s', args = %#v", name, args)
	return MakeHash(args, name, env)
}

func BaseTypeConstructorFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	P("in base type ctor, args = '%#v'", args)
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	P("in base ctor, name = '%s', args = %#v", name, args)

	return SexpNull, nil
}

func baseConstruct(env *Glisp, f *RegisteredType, nargs int) (Sexp, error) {
	if nargs > 1 {
		return SexpNull, fmt.Errorf("%d is too many arguments for a base type constructor", nargs)
	}

	v, err := f.Factory(env)
	if err != nil {
		return SexpNull, err
	}
	Q("see call to baseConstruct, v = %v/type=%T", v, v)
	if nargs == 0 {
		switch v.(type) {
		case *int, *uint8, *uint16, *uint32, *uint64, *int8, *int16, *int32, *int64:
			return SexpInt(0), nil
		case *float32, *float64:
			return SexpFloat(0), nil
		case *string:
			return SexpStr{S: ""}, nil
		case *bool:
			return SexpBool(false), nil
		case *time.Time:
			return SexpTime{}, nil
		default:
			return SexpNull, fmt.Errorf("unhandled no-arg case in baseConstruct, v has type=%T")
		}
	}

	// get our one argument
	args, err := env.datastack.PopExpressions(1)
	if err != nil {
		return SexpNull, err
	}
	arg := args[0]

	switch v.(type) {
	case *int, *uint8, *uint16, *uint32, *uint64, *int8, *int16, *int32, *int64:
		myint, ok := arg.(SexpInt)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to int", arg)
		}
		return myint, nil
	case *float32, *float64:
		myfloat, ok := arg.(SexpFloat)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to float", arg)
		}
		return myfloat, nil
	case *string:
		mystring, ok := arg.(SexpStr)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to string", arg)
		}
		return mystring, nil
	case *bool:
		mybool, ok := arg.(SexpBool)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to bool", arg)
		}
		return mybool, nil
	default:
		return SexpNull, fmt.Errorf("unhandled case in baseConstruct, arg = %#v/type=%T", arg, arg)
	}
	return SexpNull, fmt.Errorf("unhandled no-arg case in baseConstruct, v has type=%T")
}
