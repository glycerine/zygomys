package zygo

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type SexpUserVarDefn struct {
	Name string
}

type RecordDefn struct {
	Name      string
	Fields    []*SexpField
	FieldType map[string]*RegisteredType
}

func NewRecordDefn() *RecordDefn {
	return &RecordDefn{
		FieldType: make(map[string]*RegisteredType),
	}
}

func (r *RecordDefn) SetName(name string) {
	r.Name = name
}
func (r *RecordDefn) SetFields(flds []*SexpField) {
	r.Fields = flds
	for _, f := range flds {
		g := (*SexpHash)(f)
		rt, err := g.HashGet(nil, f.KeyOrder[0])
		panicOn(err)
		r.FieldType[g.KeyOrder[0].(SexpSymbol).name] = rt.(*RegisteredType)
	}
}

func (p *RecordDefn) Type() *RegisteredType {
	rt := GoStructRegistry.Registry[p.Name]
	//Q("RecordDefn) Type() sees rt = %v", rt)
	return rt
}

// pretty print a struct
func (p *RecordDefn) SexpString() string {
	Q("RecordDefn::SexpString() called!")
	if len(p.Fields) == 0 {
		return fmt.Sprintf("(struct %s)", p.Name)
	}
	s := fmt.Sprintf("(struct %s [\n", p.Name)

	w := make([][]int, len(p.Fields))
	maxnfield := 0
	for i, f := range p.Fields {
		w[i] = f.FieldWidths()
		Q("w[i=%v] = %v", i, w[i])
		maxnfield = maxi(maxnfield, len(w[i]))
	}

	biggestCol := make([]int, maxnfield)
	Q("\n")
	for j := 0; j < maxnfield; j++ {
		for i := range p.Fields {
			Q("i= %v, j=%v, len(w[i])=%v  check=%v", i, j, len(w[i]), len(w[i]) < j)
			if j < len(w[i]) {
				biggestCol[j] = maxi(biggestCol[j], w[i][j]+1)
			}
		}
	}
	Q("RecordDefn::SexpString(): maxnfield = %v, out of %v", maxnfield, len(p.Fields))
	Q("RecordDefn::SexpString(): biggestCol =  %#v", biggestCol)

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
	Q("pad = %#v", biggestCol)
	for _, f := range p.Fields {
		s += " " + f.AlignString(biggestCol) + "\n"
	}
	s += " ])\n"
	return s
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type SexpField SexpHash

func (r SexpField) Type() *RegisteredType {
	return r.GoStructFactory
}

// compute key and value widths to assist alignment
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
			wide = append(wide, len(str))
			wide = append(wide, len(val.SexpString())+1)
		} else {
			panic(err)
		}
	}
	return wide
}

func (f *SexpField) AlignString(pad []int) string {
	hash := (*SexpHash)(f)
	str := " (" + hash.TypeName + " "
	spc := " "
	for i, key := range hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		r := ""
		if err == nil {
			switch s := key.(type) {
			case SexpStr:
				r += s.S + ":"
			case SexpSymbol:
				r += s.name + ":"
			default:
				r += key.SexpString() + ":"
			}
			xtra := pad[i*2] - len(r)
			if xtra < 0 {
				panic(fmt.Sprintf("xtra = %d, pad[i=%v]=%v, len(r)=%v (r=%v)", xtra, i, pad[i], len(r), r))
			}
			leftpad := strings.Repeat(" ", xtra)
			vs := val.SexpString()
			rightpad := strings.Repeat(" ", pad[(i*2)+1]-len(vs))
			if i == 0 {
				spc = " "
			} else {
				spc = ""
			}
			r = leftpad + r + spc + vs + rightpad
		} else {
			panic(err)
		}
		str += r
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
	env.AddBuilder("var", VarBuilder)
	env.AddBuilder("expect-error", ExpectErrorBuilder)

	env.AddFunction("slice-of", SliceOfFunction)
	env.AddFunction("ptr", PointerToFunction)
}

var sxSliceOf *SexpFunction = MakeUserFunction("slice-of", SliceOfFunction)
var sxArrayOf *SexpFunction = MakeUserFunction("array-of", ArrayOfFunction)

func StructBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	n := len(args)
	if n < 1 {
		return SexpNull, fmt.Errorf("struct name is missing. use: " +
			"(struct struct-name ...)\n")
	}

	Q("in struct builder, name = '%s', args = ", name)
	for i := range args {
		Q("args[%v] = '%s' of type %T", i, args[i].SexpString(), args[i])
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
		Q("done with sn eval")
		symN, isSym := sn.(SexpSymbol)
		if !isSym {
			return SexpNull, fmt.Errorf("bad struct name: symbol required")
		}
	*/
	Q("good: have struct name '%v'", symN)

	env.datastack.PushExpr(SexpNull)
	structName := symN.name

	{
		// begin enable recursion -- add ourselves to the env early, then
		// update later, so that structs can refer to themselves.
		udsR := NewRecordDefn()
		udsR.SetName(structName)
		rtR := NewRegisteredType(func(env *Glisp) (interface{}, error) {
			return udsR, nil
		})
		rtR.UserStructDefn = udsR
		rtR.DisplayAs = structName
		GoStructRegistry.RegisterUserdef(structName, rtR, false)

		err := env.LexicalBindSymbol(symN, rtR)
		if err != nil {
			return SexpNull, fmt.Errorf("struct builder could not bind symbol '%s': '%v'",
				structName, err)
		}
		// end enable recursion
	}
	var xar []Sexp
	var flat []*SexpField
	if n > 2 {
		return SexpNull, fmt.Errorf("bad struct declaration: more than two arguments." +
			"prototype is (struct name [(field ...)*] )")
	}
	if n == 2 {
		Q("in case n == 2")
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
					Q("about to eval i=%v", i)
					ev, err := dup.EvalExpressions([]Sexp{ele})
					Q("done with eval i=%v. ev=%v", i, ev.SexpString())
					if err != nil {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field at array entry %v; error was '%v'", structName, i, err)
					}
					Q("checking for isHash at i=%v", i)
					asHash, isHash := ev.(*SexpField)
					if !isHash {
						Q("was not hash, instead was %T", ev)
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; a (field ...) is required. Instead saw '%T'/with value = '%v'",
							structName, i, ev, ev.SexpString())
					}
					Q("good eval i=%v, ev=%#v / %v", i, ev, ev.SexpString())
					ko := asHash.KeyOrder
					if len(ko) == 0 {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; field had no name",
							structName, i)
					}
					Q("ko = '%#v'", ko)
					first := ko[0]
					Q("first = '%#v'", first)
					xar = append(xar, first)
					xar = append(xar, ev)
					flat = append(flat, ev.(*SexpField))
				}
				Q("no err from EvalExpressions, got xar = '%#v'", xar)
			}
		}
		/*
				Q("evaluating args[1:2] which is of type %T / val=%#v", args[1], args[1])
				arr, err := env.EvalExpressions(args[1:2])
				if err != nil {
					return SexpNull, fmt.Errorf("bad struct declaration: bad "+
						"array of fields, error was '%v'", err)
				}

			switch ar := arr.(type) {
			case SexpArray:
				Q("good, have array %#v", ar)
				xar = []Sexp(ar)
			default:
				return SexpNull, fmt.Errorf("bad struct declaration: did not find "+
					"array of fields following name; instead found %v/type=%T", ar, ar)
			}
		*/
	} // end n == 2

	uds := NewRecordDefn()
	uds.SetName(structName)
	uds.SetFields(flat)
	Q("good: made typeDefnHash: '%s'", uds.SexpString())
	rt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return uds, nil
	})
	rt.UserStructDefn = uds
	rt.DisplayAs = structName
	GoStructRegistry.RegisterUserdef(structName, rt, false)
	Q("good: registered new userdefined struct '%s'", structName)
	err := env.linearstack.DeleteSymbolFromTopOfStackScope(symN)
	if err != nil {
		return SexpNull, fmt.Errorf("internal error: should have already had symbol '%s' "+
			"bound, but DeleteSymbolFromTopOfStackScope returned error: '%v'",
			symN.name, err)
	}
	err = env.LexicalBindSymbol(symN, rt)
	if err != nil {
		return SexpNull, fmt.Errorf("late: struct builder could not bind symbol '%s': '%v'",
			structName, err)
	}
	Q("good: bound symbol '%s' to RegisteredType '%s'", symN.SexpString(), rt.SexpString())
	return rt, nil
}

// this is just a stub. TODO: finish design, implement packages.
func PackageBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("package name is missing. use: " +
			"(package package-name ...)\n")
	}

	Q("in package builder, args = ")
	for i := range args {
		Q("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}

func InterfaceBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("interface name is missing. use: " +
			"(interface interface-name ...)\n")
	}

	Q("in interface builder, args = ")
	for i := range args {
		Q("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}

func FuncBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("func name is missing. use: " +
			"(func func-name ...)\n")
	}

	Q("in func builder, args = ")
	for i := range args {
		Q("args[%v] = '%s'", i, args[i].SexpString())
	}

	return SexpNull, nil
}

func SliceOfFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument x to (%s x) is missing. use: "+
			"(%s a-regtype)\n", name)
	}

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	default:
		return SexpNull, fmt.Errorf("argument tx in (%s x) was not regtype, "+
			"instead type %T displaying as '%v' ",
			name, arg, arg.SexpString())
	}

	//Q("slice-of arg = '%s' with type %T", args[0].SexpString(), args[0])

	derivedType := reflect.SliceOf(rt.TypeCache)
	sliceRt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return reflect.MakeSlice(derivedType, 0, 0), nil
	})
	sliceRt.DisplayAs = fmt.Sprintf("(%s %s)", name, rt.DisplayAs)
	sliceName := "slice-of-" + rt.RegisteredName
	GoStructRegistry.RegisterUserdef(sliceName, sliceRt, false)
	return sliceRt, nil
}

func PointerToFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument to pointer-to is missing. use: "+
			"(%s a-regtype)\n", name)
	}

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	case *SexpPointer:
		// dereference operation, rather than type declaration
		Q("dereference operation on *SexpPointer detected, returning target")
		if arg == nil || arg.Target == nil {
			return SexpNull, fmt.Errorf("illegal to dereference nil pointer")
		}
		return arg.Target, nil
	case SexpReflect:
		Q("dereference operation on SexpReflect detected")
		// TODO what goes here?
		return SexpNull, fmt.Errorf("illegal to dereference nil pointer")
	default:
		return SexpNull, fmt.Errorf("argument x in (%s x) was not regtype or SexpPointer, "+
			"instead type %T displaying as '%v' ",
			name, arg, arg.SexpString())
	}

	//Q("pointer-to arg = '%s' with type %T", args[0].SexpString(), args[0])

	ptrRt := GoStructRegistry.GetOrCreatePointerType(rt)
	return ptrRt, nil
}

func StructConstructorFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	Q("in struct ctor, name = '%s', args = %#v", name, args)
	return MakeHash(args, name, env)
}

func BaseTypeConstructorFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	Q("in base type ctor, args = '%#v'", args)
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	Q("in base ctor, name = '%s', args = %#v", name, args)

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
			return SexpInt{}, nil
		case *float32, *float64:
			return SexpFloat{}, nil
		case *string:
			return SexpStr{S: ""}, nil
		case *bool:
			return SexpBool{Val: false}, nil
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

// generate fixed size array
func ArrayOfFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 2 {
		return SexpNull, fmt.Errorf("insufficient arguments to ([size] regtype) array constructor. use: "+
			"([size...] a-regtype)\n", name)
	}
	sz := 0
	Q("args = %#v in ArrayOfFunction", args)
	switch ar := args[1].(type) {
	case SexpArray:
		if len(ar) == 0 {
			return SexpNull, fmt.Errorf("at least one size must be specified in array constructor; e.g. ([size ...] regtype)")
		}
		asInt, isInt := ar[0].(SexpInt)
		if !isInt {
			return SexpNull, fmt.Errorf("size must be an int (not %T) in array constructor; e.g. ([size ...] regtype)", ar[0])
		}
		sz = int(asInt.Val)
		// TODO: implement multiple dimensional arrays (matrixes etc).
	default:
		return SexpNull, fmt.Errorf("at least one size must be specified in array constructor; e.g. ([size ...] regtype)")
	}

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	default:
		return SexpNull, fmt.Errorf("argument tx in (%s x) was not regtype, "+
			"instead type %T displaying as '%v' ",
			name, arg, arg.SexpString())
	}

	//Q("array-of arg = '%s' with type %T", args[0].SexpString(), args[0])

	derivedType := reflect.ArrayOf(sz, rt.TypeCache)
	arrayRt := NewRegisteredType(func(env *Glisp) (interface{}, error) {
		return reflect.New(derivedType), nil
	})
	arrayRt.DisplayAs = fmt.Sprintf("(%s %s)", name, rt.DisplayAs)
	arrayName := "array-of-" + rt.RegisteredName
	GoStructRegistry.RegisterUserdef(arrayName, arrayRt, false)
	return arrayRt, nil
}

func VarBuilder(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	n := len(args)
	if n != 2 {
		return SexpNull, fmt.Errorf("var name/type missing. use: " +
			"(var name regtype)\n")
	}

	Q("in var builder, name = '%s', args = ", name)
	for i := range args {
		Q("args[%v] = '%s' of type %T", i, args[i].SexpString(), args[i])
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
			return SexpNull, fmt.Errorf("bad var name: symbol required")
		}
	default:
		return SexpNull, fmt.Errorf("bad var name: symbol required")
	}
	Q("good: have var name '%v'", symN)

	dup := env.Duplicate()
	Q("about to eval args[1]=%v", args[1])
	ev, err := dup.EvalExpressions(args[1:2])
	Q("done with eval, ev=%v / type %T", ev.SexpString(), ev)
	if err != nil {
		return SexpNull, fmt.Errorf("bad var declaration, problem with type '%v': %v", args[1].SexpString(), err)
	}

	var rt *RegisteredType
	switch myrt := ev.(type) {
	case *RegisteredType:
		rt = myrt
	default:
		return SexpNull, fmt.Errorf("bad var declaration, type '%v' is unknown", rt.SexpString())
	}

	val, err := rt.Factory(env)
	if err != nil {
		return SexpNull, fmt.Errorf("var declaration error: could not make type '%s': %v",
			rt, err)
	}
	var valSexp Sexp
	Q("val is of type %T", val)
	switch v := val.(type) {
	case Sexp:
		valSexp = v
	case reflect.Value:
		Q("v is of type %T", v.Interface())
		switch rd := v.Interface().(type) {
		case ***RecordDefn:
			Q("we have RecordDefn rd = %#v", *rd)
		}
		valSexp = SexpReflect(reflect.ValueOf(v))
	default:
		valSexp = SexpReflect(reflect.ValueOf(v))
	}

	Q("var decl: valSexp is '%v'", valSexp.SexpString())
	err = env.LexicalBindSymbol(symN, valSexp)
	if err != nil {
		return SexpNull, fmt.Errorf("var declaration error: could not bind symbol '%s': %v",
			symN.name, err)
	}
	Q("good: var decl bound symbol '%s' to '%s' of type '%s'", symN.SexpString(), valSexp.SexpString(), rt.SexpString())

	env.datastack.PushExpr(valSexp)

	return SexpNull, nil
}

func ExpectErrorBuilder(env *Glisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 2 {
		return SexpNull, WrongNargs
	}

	var expectedError SexpStr
	switch e := args[0].(type) {
	case SexpStr:
		expectedError = e
	default:
		return SexpNull, fmt.Errorf("first arg to expect-error must be the string of the error to expect")
	}
	Q("expectedError = %v", expectedError)
	dup := env.Duplicate()
	ev, err := dup.EvalExpressions(args[1:2])
	Q("done with eval, ev=%v / type %T. err = %v", ev.SexpString(), ev, err)
	if err != nil {
		if err.Error() == expectedError.S {
			return SexpNull, nil
		}
		return SexpNull, fmt.Errorf("expect-error expected '%s' but saw '%s'", expectedError.S, err)
	}

	if expectedError.S == "" {
		return SexpNull, nil
	}
	return SexpNull, fmt.Errorf("expect-error expected '%s' but got no error", expectedError.S)
}
