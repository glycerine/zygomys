package zygo

import (
	"fmt"
	"reflect"
	"strings"
	"time"
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
// Since it receives an un-evaluated tree of
// symbols, a builder must manually evaluate
// any arguments it wishes to find bindings for.
//
// The primary use here is to be able to define
// packages, structs, interfaces, functions,
// methods, and type aliases.
func (env *Zlisp) ImportPackageBuilder() {
	env.AddBuilder("infixExpand", InfixBuilder)
	env.AddBuilder("infix", InfixBuilder)
	env.AddBuilder(":", ColonAccessBuilder)
	env.AddBuilder("sys", SystemBuilder)
	env.AddBuilder("struct", StructBuilder)
	env.AddBuilder("func", FuncBuilder)
	env.AddBuilder("method", FuncBuilder)
	env.AddBuilder("interface", InterfaceBuilder)
	//env.AddBuilder("package", PackageBuilder)
	//env.AddBuilder("import", ImportBuilder)
	env.AddBuilder("var", VarBuilder)
	env.AddBuilder("expectError", ExpectErrorBuilder)
	env.AddBuilder("comma", CommaBuilder)
	env.AddBuilder("raw64", Raw64Builder)
	//	env.AddBuilder("&", AddressOfBuilder)

	env.AddBuilder("import", ImportPackageBuilder)

	env.AddFunction("sliceOf", SliceOfFunction)
	env.AddFunction("ptr", PointerToFunction)
}

var sxSliceOf *SexpFunction = MakeUserFunction("sliceOf", SliceOfFunction)
var sxArrayOf *SexpFunction = MakeUserFunction("arrayOf", ArrayOfFunction)

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
		r.FieldType[g.KeyOrder[0].(*SexpSymbol).name] = rt.(*RegisteredType)
	}
}

func (p *RecordDefn) Type() *RegisteredType {
	rt := GoStructRegistry.Registry[p.Name]
	//Q("RecordDefn) Type() sees rt = %v", rt)
	return rt
}

// pretty print a struct
func (p *RecordDefn) SexpString(ps *PrintState) string {
	//Q("RecordDefn::SexpString() called!")
	if len(p.Fields) == 0 {
		return fmt.Sprintf("(struct %s)", p.Name)
	}
	s := fmt.Sprintf("(struct %s [\n", p.Name)

	w := make([][]int, len(p.Fields))
	maxnfield := 0
	for i, f := range p.Fields {
		w[i] = f.FieldWidths()
		//Q("w[i=%v] = %v", i, w[i])
		maxnfield = maxi(maxnfield, len(w[i]))
	}

	biggestCol := make([]int, maxnfield)
	//Q("\n")
	for j := 0; j < maxnfield; j++ {
		for i := range p.Fields {
			//Q("i= %v, j=%v, len(w[i])=%v  check=%v", i, j, len(w[i]), len(w[i]) < j)
			if j < len(w[i]) {
				biggestCol[j] = maxi(biggestCol[j], w[i][j]+1)
			}
		}
	}
	//Q("RecordDefn::SexpString(): maxnfield = %v, out of %v", maxnfield, len(p.Fields))
	//Q("RecordDefn::SexpString(): biggestCol =  %#v", biggestCol)

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
	//Q("pad = %#v", biggestCol)
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
			case *SexpStr:
				str += s.S + ":"
			case *SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString(nil) + ":"
			}
			wide = append(wide, len(str))
			wide = append(wide, len(val.SexpString(nil))+1)
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
			case *SexpStr:
				r += s.S + ":"
			case *SexpSymbol:
				r += s.name + ":"
			default:
				r += key.SexpString(nil) + ":"
			}
			xtra := pad[i*2] - len(r)
			if xtra < 0 {
				panic(fmt.Sprintf("xtra = %d, pad[i=%v]=%v, len(r)=%v (r=%v)", xtra, i, pad[i], len(r), r))
			}
			leftpad := strings.Repeat(" ", xtra)
			vs := val.SexpString(nil)
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

func (f *SexpField) SexpString(ps *PrintState) string {
	hash := (*SexpHash)(f)
	str := " (" + hash.TypeName + " "

	for i, key := range hash.KeyOrder {
		val, err := hash.HashGet(nil, key)
		if err == nil {
			switch s := key.(type) {
			case *SexpStr:
				str += s.S + ":"
			case *SexpSymbol:
				str += s.name + ":"
			default:
				str += key.SexpString(nil) + ":"
			}
			if i > 0 {
				str += val.SexpString(nil) + " "
			} else {
				str += val.SexpString(nil) + "    "
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

func StructBuilder(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	n := len(args)
	if n < 1 {
		return SexpNull, fmt.Errorf("struct name is missing. use: " +
			"(struct struct-name ...)\n")
	}

	//Q("in struct builder, name = '%s', args = ", name)
	//for i := range args {
	//Q("args[%v] = '%s' of type %T", i, args[i].SexpString(nil), args[i])
	//}
	var symN *SexpSymbol
	switch b := args[0].(type) {
	case *SexpSymbol:
		symN = b
	case *SexpPair:
		sy, isQuo := isQuotedSymbol(b)
		if isQuo {
			symN = sy.(*SexpSymbol)
		} else {
			return SexpNull, fmt.Errorf("bad struct name: symbol required")
		}
	default:
		return SexpNull, fmt.Errorf("bad struct name: symbol required")
	}

	//Q("good: have struct name '%v'", symN)

	env.datastack.PushExpr(SexpNull)
	structName := symN.name

	{
		// begin enable recursion -- add ourselves to the env early, then
		// update later, so that structs can refer to themselves.
		udsR := NewRecordDefn()
		udsR.SetName(structName)
		rtR := NewRegisteredType(func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return udsR, nil
		})
		rtR.UserStructDefn = udsR
		rtR.DisplayAs = structName
		GoStructRegistry.RegisterUserdef(rtR, false, structName)

		// overwrite any existing definition, deliberately ignore any error,
		// as there may not be a prior definition present at all.
		env.linearstack.DeleteSymbolFromTopOfStackScope(symN)

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
		//Q("in case n == 2")
		switch ar := args[1].(type) {
		default:
			return SexpNull, fmt.Errorf("bad struct declaration '%v': second argument "+
				"must be a slice of fields."+
				" prototype is (struct name [(field ...)*] )", structName)
		case *SexpArray:
			arr := ar.Val
			if len(arr) == 0 {
				// allow this
			} else {
				// dup to avoid messing with the stack on eval:
				//dup := env.Duplicate()
				for i, ele := range arr {
					//Q("about to eval i=%v", i)
					//ev, err := dup.EvalExpressions([]Sexp{ele})
					ev, err := EvalFunction(env, "evalStructBuilder", []Sexp{ele})
					//Q("done with eval i=%v. ev=%v", i, ev.SexpString(nil))
					if err != nil {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field at array entry %v; error was '%v'", structName, i, err)
					}
					//Q("checking for isHash at i=%v", i)
					asHash, isHash := ev.(*SexpField)
					if !isHash {
						//Q("was not hash, instead was %T", ev)
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; a (field ...) is required. Instead saw '%T'/with value = '%v'",
							structName, i, ev, ev.SexpString(nil))
					}
					//Q("good eval i=%v, ev=%#v / %v", i, ev, ev.SexpString(nil))
					ko := asHash.KeyOrder
					if len(ko) == 0 {
						return SexpNull, fmt.Errorf("bad struct declaration '%v': bad "+
							"field array at entry %v; field had no name",
							structName, i)
					}
					//Q("ko = '%#v'", ko)
					first := ko[0]
					//Q("first = '%#v'", first)
					xar = append(xar, first)
					xar = append(xar, ev)
					flat = append(flat, ev.(*SexpField))
				}
				//Q("no err from EvalExpressions, got xar = '%#v'", xar)
			}
		}
	} // end n == 2

	uds := NewRecordDefn()
	uds.SetName(structName)
	uds.SetFields(flat)
	//Q("good: made typeDefnHash: '%s'", uds.SexpString(nil))
	rt := NewRegisteredType(func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return uds, nil
	})
	rt.UserStructDefn = uds
	rt.DisplayAs = structName
	GoStructRegistry.RegisterUserdef(rt, false, structName)
	//Q("good: registered new userdefined struct '%s'", structName)

	// replace our recursive-reference-enabling symbol with the real one.
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
	//Q("good: bound symbol '%s' to RegisteredType '%s'", symN.SexpString(nil), rt.SexpString(nil))
	return rt, nil
}

func InterfaceBuilder(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	nargs := len(args)
	switch {
	case nargs < 1:
		return SexpNull, fmt.Errorf("interface name is missing. use: " +
			"(interface interface-name [...])\n")
	case nargs == 1:
		return SexpNull, fmt.Errorf("interface array of methods missing. use: " +
			"(interface interface-name [...])\n")
	case nargs > 2:
		return SexpNull, WrongNargs
	}

	//	P("in interface builder, past arg check")
	var iname string
	var symN *SexpSymbol
	switch sy := args[0].(type) {
	case *SexpSymbol:
		symN = sy
		iname = sy.name
	default:
		return SexpNull, fmt.Errorf("interface name must be a symbol; we got %T", args[0])
	}

	// sanity check the name
	builtin, builtTyp := env.IsBuiltinSym(symN)
	if builtin {
		return SexpNull,
			fmt.Errorf("already have %s '%s', refusing to overwrite with interface",
				builtTyp, symN.name)
	}

	if env.HasMacro(symN) {
		return SexpNull, fmt.Errorf("Already have macro named '%s': refusing"+
			" to define interface  of same name.", symN.name)
	}
	// end sanity check the name

	var arrMeth *SexpArray
	switch ar := args[1].(type) {
	case *SexpArray:
		arrMeth = ar
	default:
		return SexpNull, fmt.Errorf("interface method vector expected after name; we got %T", args[1])
	}

	//	P("in interface builder, args = ")
	//	for i := range args {
	//		P("args[%v] = '%s'", i, args[i].SexpString(nil))
	//	}

	methods := make([]*SexpFunction, 0)
	methodSlice := arrMeth.Val
	if len(methodSlice) > 0 {
		//dup := env.Duplicate()
		for i := range methodSlice {
			//ev, err := dup.EvalExpressions([]Sexp{methodSlice[i]})
			ev, err := EvalFunction(env, "evalInterface", []Sexp{methodSlice[i]})
			if err != nil {
				return SexpNull, fmt.Errorf("error parsing the %v-th method in interface definition: '%v'", i, err)
			}
			me, gotFunc := ev.(*SexpFunction)
			if !gotFunc {
				return SexpNull,
					fmt.Errorf("error parsing the %v-th method in interface: was not function but rather %T",
						i, ev)
			}
			methods = append(methods, me)
		}
	}

	decl := &SexpInterfaceDecl{
		name:    iname,
		methods: methods,
	}
	return decl, nil
}

func SliceOfFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument x to (%s x) is missing. use: "+
			"(%s a-regtype)\n", name, name)
	}

	//Q("in SliceOfFunction")

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	default:
		return SexpNull, fmt.Errorf("argument tx in (%s x) was not regtype, "+
			"instead type %T displaying as '%v' ",
			name, arg, arg.SexpString(nil))
	}

	//Q("sliceOf arg = '%s' with type %T", args[0].SexpString(nil), args[0])

	sliceRt := GoStructRegistry.GetOrCreateSliceType(rt)
	//Q("in SliceOfFunction: returning sliceRt = '%#v'", sliceRt)
	return sliceRt, nil
}

func PointerToFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, fmt.Errorf("argument to pointer-to is missing. use: "+
			"(%s a-regtype)\n", name)
	}

	//P("in PointerToFunction(): args[0] = '%#v'", args[0])

	var rt *RegisteredType
	switch arg := args[0].(type) {
	case *RegisteredType:
		rt = arg
	case *SexpHash:
		rt = arg.GoStructFactory
	case *SexpPointer:
		// dereference operation, rather than type declaration
		//P("dereference operation on *SexpPointer detected, returning target")
		if arg == nil || arg.Target == nil {
			return SexpNull, fmt.Errorf("illegal to dereference nil pointer")
		}
		return arg.Target, nil
	case *SexpReflect:
		//Q("dereference operation on SexpReflect detected")
		// TODO what goes here?
		return SexpNull, fmt.Errorf("illegal to dereference nil pointer")
	case *SexpSymbol:
		if arg.isDot {
			// (* h.a) dereferencing a dot symbol
			resolved, err := dotGetSetHelper(env, arg.name, nil)
			if err != nil {
				return nil, err
			}
			return resolved, nil
		} else {
			panic("TODO: what goes here, for (* sym) where sym is a regular symbol")
		}
	default:
		return SexpNull, fmt.Errorf("argument x in (%s x) was not regtype or SexpPointer, "+
			"instead type %T displaying as '%v' ",
			name, arg, arg.SexpString(nil))
	}

	//Q("pointer-to arg = '%s' with type %T", args[0].SexpString(nil), args[0])

	ptrRt := GoStructRegistry.GetOrCreatePointerType(rt)
	return ptrRt, nil
}

func StructConstructorFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//Q("in struct ctor, name = '%s', args = %#v", name, args)
	return MakeHash(args, name, env)
}

func BaseTypeConstructorFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//Q("in base type ctor, args = '%#v'", args)
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	//Q("in base ctor, name = '%s', args = %#v", name, args)

	return SexpNull, nil
}

func baseConstruct(env *Zlisp, f *RegisteredType, nargs int) (Sexp, error) {
	if nargs > 1 {
		return SexpNull, fmt.Errorf("%d is too many arguments for a base type constructor", nargs)
	}

	v, err := f.Factory(env, nil)
	if err != nil {
		return SexpNull, err
	}
	//Q("see call to baseConstruct, v = %v/type=%T", v, v)
	if nargs == 0 {
		switch v.(type) {
		case *int, *uint8, *uint16, *uint32, *uint64, *int8, *int16, *int32, *int64:
			return &SexpInt{}, nil
		case *float32, *float64:
			return &SexpFloat{}, nil
		case *string:
			return &SexpStr{S: ""}, nil
		case *bool:
			return &SexpBool{Val: false}, nil
		case *time.Time:
			return &SexpTime{}, nil
		default:
			return SexpNull, fmt.Errorf("unhandled no-arg case in baseConstruct, v has type=%T", v)
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
		myint, ok := arg.(*SexpInt)
		if !ok {
			// try from Float, even though it will loose the decimals.
			myfl, ok2 := arg.(*SexpFloat)
			if ok2 {
				return &SexpInt{Val: int64(myfl.Val)}, nil
			}
			return SexpNull, fmt.Errorf("cannot convert %T to int", arg)
		}
		return myint, nil
	case *float32, *float64:
		myfloat, ok := arg.(*SexpFloat)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to float", arg)
		}
		return myfloat, nil
	case *string:
		mystring, ok := arg.(*SexpStr)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to string", arg)
		}
		return mystring, nil
	case *bool:
		mybool, ok := arg.(*SexpBool)
		if !ok {
			return SexpNull, fmt.Errorf("cannot convert %T to bool", arg)
		}
		return mybool, nil
	default:
		return SexpNull, fmt.Errorf("unhandled case in baseConstruct, arg = %#v/type=%T", arg, arg)
	}
	//return SexpNull, fmt.Errorf("unhandled no-arg case in baseConstruct, v has type=%T", v)
}

// generate fixed size array
func ArrayOfFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 2 {
		return SexpNull, fmt.Errorf("insufficient arguments to ([size] regtype) array constructor. use: " +
			"([size...] a-regtype)\n")
	}
	sz := 0
	//Q("args = %#v in ArrayOfFunction", args)
	switch ar := args[1].(type) {
	case *SexpArray:
		if len(ar.Val) == 0 {
			return SexpNull, fmt.Errorf("at least one size must be specified in array constructor; e.g. ([size ...] regtype)")
		}
		asInt, isInt := ar.Val[0].(*SexpInt)
		if !isInt {
			return SexpNull, fmt.Errorf("size must be an int (not %T) in array constructor; e.g. ([size ...] regtype)", ar.Val[0])
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
			name, arg, arg.SexpString(nil))
	}

	//Q("arrayOf arg = '%s' with type %T", args[0].SexpString(nil), args[0])

	derivedType := reflect.ArrayOf(sz, rt.TypeCache)
	arrayRt := NewRegisteredType(func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return reflect.New(derivedType), nil
	})
	arrayRt.DisplayAs = fmt.Sprintf("(%s %s)", name, rt.DisplayAs)
	arrayName := "arrayOf" + rt.RegisteredName
	GoStructRegistry.RegisterUserdef(arrayRt, false, arrayName)
	return arrayRt, nil
}

func VarBuilder(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	n := len(args)
	if n != 2 {
		return SexpNull, fmt.Errorf("var name/type missing. use: " +
			"(var name regtype)\n")
	}

	//Q("in var builder, name = '%s', args = ", name)
	//for i := range args {
	//Q("args[%v] = '%s' of type %T", i, args[i].SexpString(nil), args[i])
	//}
	var symN *SexpSymbol
	switch b := args[0].(type) {
	case *SexpSymbol:
		symN = b
	case *SexpPair:
		sy, isQuo := isQuotedSymbol(b)
		if isQuo {
			symN = sy.(*SexpSymbol)
		} else {
			return SexpNull, fmt.Errorf("bad var name: symbol required")
		}
	default:
		return SexpNull, fmt.Errorf("bad var name: symbol required")
	}
	//Q("good: have var name '%v'", symN)

	//dup := env.Duplicate()
	//Q("about to eval args[1]=%v", args[1])
	//ev, err := dup.EvalExpressions(args[1:2])
	ev, err := EvalFunction(env, "evalVar", args[1:2])
	//Q("done with eval, ev=%v / type %T", ev.SexpString(nil), ev)
	if err != nil {
		return SexpNull, fmt.Errorf("bad var declaration, problem with type '%v': %v", args[1].SexpString(nil), err)
	}

	var rt *RegisteredType
	switch myrt := ev.(type) {
	case *RegisteredType:
		rt = myrt
	default:
		return SexpNull, fmt.Errorf("bad var declaration, type '%v' is unknown", rt.SexpString(nil))
	}

	val, err := rt.Factory(env, nil)
	if err != nil {
		return SexpNull, fmt.Errorf("var declaration error: could not make type '%s': %v",
			rt.SexpString(nil), err)
	}
	var valSexp Sexp
	//Q("val is of type %T", val)
	switch v := val.(type) {
	case Sexp:
		valSexp = v
	case reflect.Value:
		//Q("v is of type %T", v.Interface())
		switch rd := v.Interface().(type) {
		case ***RecordDefn:
			_ = rd
			//Q("we have RecordDefn rd = %#v", *rd)
		}
		valSexp = &SexpReflect{Val: reflect.ValueOf(v)}
	default:
		valSexp = &SexpReflect{Val: reflect.ValueOf(v)}
	}

	//Q("var decl: valSexp is '%v'", valSexp.SexpString(nil))
	err = env.LexicalBindSymbol(symN, valSexp)
	if err != nil {
		return SexpNull, fmt.Errorf("var declaration error: could not bind symbol '%s': %v",
			symN.name, err)
	}
	//Q("good: var decl bound symbol '%s' to '%s' of type '%s'", symN.SexpString(nil), valSexp.SexpString(nil), rt.SexpString(nil))

	env.datastack.PushExpr(valSexp)

	return SexpNull, nil
}

func ExpectErrorBuilder(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 2 {
		return SexpNull, WrongNargs
	}

	dup := env.Duplicate()
	es, err := dup.EvalExpressions(args[0:1])
	if err != nil {
		return SexpNull, fmt.Errorf("error evaluating the error string to expect: %s", err)
	}

	var expectedError *SexpStr
	switch e := es.(type) {
	case *SexpStr:
		expectedError = e
	default:
		return SexpNull, fmt.Errorf("first arg to expectError must be the string of the error to expect")
	}
	//Q("expectedError = %v", expectedError)
	ev, err := dup.EvalExpressions(args[1:2])
	_ = ev
	//Q("done with eval, ev=%v / type %T. err = %v", ev.SexpString(nil), ev, err)
	if err != nil {
		if err.Error() == expectedError.S {
			return SexpNull, nil
		}
		return SexpNull, fmt.Errorf("expectError expected '%s' but saw '%s'", expectedError.S, err)
	}

	if expectedError.S == "" {
		return SexpNull, nil
	}
	return SexpNull, fmt.Errorf("expectError expected '%s' but got no error", expectedError.S)
}

func ColonAccessBuilder(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return SexpNull, WrongNargs
	}
	////Q("ColonAccessBuilder, args = %#v", args)
	name = "hget"

	//dup := env.Duplicate()
	//collec, err := dup.EvalExpressions(args[1:2])
	collec, err := EvalFunction(env, "evalColonAccess", args[1:2])
	if err != nil {
		return SexpNull, err
	}
	swapped := args
	swapped[1] = swapped[0]
	swapped[0] = collec

	if len(args) == 3 {
		// have default, needs eval too
		//defaul, err := dup.EvalExpressions(args[2:3])
		defaul, err := EvalFunction(env, "evalColonDefault", args[2:3])
		if err != nil {
			return SexpNull, err
		}
		swapped[2] = defaul
	}

	switch sx := collec.(type) {
	case *SexpHash:
		return HashAccessFunction(name)(env, name, swapped)
	case *SexpArray:
		return ArrayAccessFunction(name)(env, name, swapped)
	case *SexpArraySelector:
		//Q("*SexpSelector seen in : operator form.")
		return sx.RHS(env)
	}
	return SexpNull, fmt.Errorf("second argument to ':' function must be hash or array")
}

// CommaBuilder turns expressions on the LHS and RHS like {a,b,c = 1,2,3}
// into arrays (set [a b c] [1 2 3])
func CommaBuilder(env *Zlisp, name string, args []Sexp) (Sexp, error) {

	n := len(args)
	if n < 1 {
		return SexpNull, nil
	}

	res := make([]Sexp, 0)
	for i := range args {
		commaHelper(args[i], &res)
	}
	return &SexpArray{Val: res}, nil
}

func commaHelper(a Sexp, res *[]Sexp) {
	//Q("top of commaHelper with a = '%s'", a.SexpString(nil))
	switch x := a.(type) {
	case *SexpPair:
		sy, isQuo := isQuotedSymbol(x)
		if isQuo {
			symN := sy.(*SexpSymbol)
			//Q("have quoted symbol symN=%v", symN.SexpString(nil))
			*res = append(*res, symN)
			return
		}

		ar, err := ListToArray(x)
		if err != nil || len(ar) < 1 {
			return
		}

		switch sym := ar[0].(type) {
		case *SexpSymbol:
			if sym.name == "comma" {
				//Q("have recursive comma")
				over := ar[1:]
				for i := range over {
					commaHelper(over[i], res)
				}
			} else {
				//Q("have symbol sym = '%v'", sym.SexpString(nil))
				*res = append(*res, a)
			}
		default:
			*res = append(*res, a)
		}
	default:
		*res = append(*res, a)
	}
}
