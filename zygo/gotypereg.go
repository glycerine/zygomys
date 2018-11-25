package zygo

import (
	"fmt"
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
	// comprehensive
	Registry map[string]*RegisteredType

	// only init-time builtins
	Builtin map[string]*RegisteredType

	// later, user-defined types
	Userdef map[string]*RegisteredType
}

// consistently ordered list of all registered types (created at init time).
var ListRegisteredTypes = []string{}

func (r *GoStructRegistryType) RegisterBuiltin(name string, e *RegisteredType) {
	r.register(name, e, false)
	e.IsUser = false
}

func (r *GoStructRegistryType) RegisterPointer(pointedToName string, pointedToType *RegisteredType) *RegisteredType {
	newRT := &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		p, err := pointedToType.Factory(env, h)
		if err != nil {
			return nil, err
		}
		return &p, nil
	}}
	r.register(fmt.Sprintf("(* %s)", pointedToName), newRT, false)
	newRT.IsPointer = true
	return newRT
}

func (r *GoStructRegistryType) register(name string, e *RegisteredType, isUser bool) {
	if !e.initDone {
		e.Init()
	}
	e.RegisteredName = name
	e.Aliases[name] = true
	e.Aliases[e.ReflectName] = true

	_, found := r.Registry[name]
	if !found {
		ListRegisteredTypes = append(ListRegisteredTypes, name)
	}
	_, found2 := r.Registry[e.ReflectName]
	if !found2 {
		ListRegisteredTypes = append(ListRegisteredTypes, e.ReflectName)
	}

	if isUser {
		r.Userdef[name] = e
	} else {
		r.Builtin[name] = e
	}
	r.Registry[name] = e
	r.Registry[e.ReflectName] = e
}

func (e *RegisteredType) Init() {
	e.Aliases = make(map[string]bool)
	val, err := e.Factory(nil, nil)
	panicOn(err)
	if val != nil {
		e.ValueCache = reflect.ValueOf(val)
		e.TypeCache = e.ValueCache.Type()
		e.PointerName = fmt.Sprintf("%T", val)
		e.ReflectName = e.PointerName[1:] // todo: make this conditional on whether PointerName starts with '*'.
		e.DisplayAs = e.ReflectName
	}
	e.initDone = true
}

func reflectName(val reflect.Value) string {
	pointerName := fmt.Sprintf("%T", val.Interface())
	reflectName := pointerName[1:]
	return reflectName
}
func ifaceName(val interface{}) string {
	pointerName := fmt.Sprintf("%T", val)
	reflectName := pointerName[1:]
	return reflectName
}

func (r *GoStructRegistryType) RegisterUserdef(
	e *RegisteredType,
	hasShadowStruct bool,
	names ...string) {

	for i, name := range names {
		e0 := e
		if i > 0 {
			// make a copy of the RegisteredType for each name, so all names are kept.
			// Otherwise we overwrite the DisplayAs below.
			rt := *e
			e0 = &rt
		}
		r.register(name, e0, true)
		e0.IsUser = true
		e0.hasShadowStruct = hasShadowStruct

		e0.Constructor = MakeUserFunction("__struct_"+name, StructConstructorFunction)
		if e0.DisplayAs == "" {
			e0.DisplayAs = name
		}
	}
}

func (r *GoStructRegistryType) Lookup(name string) *RegisteredType {
	return r.Registry[name]
}

// the type of all maker functions

type MakeGoStructFunc func(env *Zlisp, h *SexpHash) (interface{}, error)

var NullRT *RegisteredType
var PairRT *RegisteredType
var Int64RT *RegisteredType
var BoolRT *RegisteredType
var RuneRT *RegisteredType
var Float64RT *RegisteredType
var RawRT *RegisteredType
var ReflectRT *RegisteredType
var ErrorRT *RegisteredType
var SentinelRT *RegisteredType
var ClosureRT *RegisteredType
var ArraySelectorRT *RegisteredType

type RegisteredType struct {
	initDone        bool
	hasShadowStruct bool

	Constructor    *SexpFunction
	RegisteredName string
	Factory        MakeGoStructFunc
	GenDefMap      bool
	ValueCache     reflect.Value
	TypeCache      reflect.Type
	PointerName    string
	ReflectName    string
	IsUser         bool
	Aliases        map[string]bool
	DisplayAs      string
	UserStructDefn *RecordDefn
	IsPointer      bool
}

func (p *RegisteredType) TypeCheckRecord(hash *SexpHash) error {
	Q("in RegisteredType.TypeCheckRecord(hash = '%v')", hash.SexpString(nil))
	if hash.TypeName == "field" {
		Q("in RegisteredType.TypeCheckRecord, TypeName == field, skipping.")
		return nil
	}
	if p.UserStructDefn != nil {
		Q("in RegisteredType.TypeCheckRecord, type checking against '%#v'", p.UserStructDefn)

		var err error
		for _, key := range hash.KeyOrder {
			obs, _ := hash.HashGet(nil, key)
			err = hash.TypeCheckField(key, obs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *RegisteredType) SexpString(ps *PrintState) string {
	if p == nil {
		return "nil RegisteredType"
	}
	if p.UserStructDefn != nil {
		return p.UserStructDefn.SexpString(ps)
	}
	return p.DisplayAs
}

func (p *RegisteredType) ShortName() string {
	if p.UserStructDefn != nil {
		return p.UserStructDefn.Name
	}
	return p.DisplayAs
}

func NewRegisteredType(f MakeGoStructFunc) *RegisteredType {
	rt := &RegisteredType{Factory: f}
	rt.Init()
	return rt
}

// builtin known Go Structs
// NB these are used to test the functionality of the
//    Go integration.
//
func init() {
	GoStructRegistry = GoStructRegistryType{
		Registry: make(map[string]*RegisteredType),
		Builtin:  make(map[string]*RegisteredType),
		Userdef:  make(map[string]*RegisteredType),
	}

	gsr := &GoStructRegistry

	// add go builtin types
	// ====================

	// empty array
	gsr.RegisterBuiltin("[]", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &SexpArray{}, nil
	}})

	// scope, as used by the package operation
	gsr.RegisterBuiltin("packageScope", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		pkg := env.NewScope()
		pkg.Name = "prototype"
		pkg.IsPackage = true
		return pkg, nil
	}})

	gsr.RegisterBuiltin("packageScopeStack", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		pkg := env.NewStack(0)
		pkg.Name = "prototypePackageScopeStack"
		pkg.IsPackage = true
		return pkg, nil
	}})

	gsr.RegisterBuiltin("arraySelector", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &SexpArraySelector{}, nil
	}})

	gsr.RegisterBuiltin("hashSelector", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &SexpHashSelector{}, nil
	}})

	gsr.RegisterBuiltin("comment",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return SexpNull, nil
		}})
	gsr.RegisterBuiltin("byte",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(byte), nil
		}})
	gsr.RegisterBuiltin("uint8",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(byte), nil
		}})

	gsr.RegisterBuiltin("int",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(int), nil
		}})
	gsr.RegisterBuiltin("uint16",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(uint16), nil
		}})
	gsr.RegisterBuiltin("uint32",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(uint32), nil
		}})
	gsr.RegisterBuiltin("uint64",
		&RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return new(uint64), nil
		}})
	gsr.RegisterBuiltin("int8", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(int8), nil
	}})
	gsr.RegisterBuiltin("int16", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(int16), nil
	}})
	gsr.RegisterBuiltin("int32", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(int32), nil
	}})
	gsr.RegisterBuiltin("rune", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(int32), nil
	}})
	gsr.RegisterBuiltin("int64", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(int64), nil
	}})
	gsr.RegisterBuiltin("float32", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(float32), nil
	}})

	gsr.RegisterBuiltin("float64", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(float64), nil
	}})

	gsr.RegisterBuiltin("complex64", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(complex64), nil
	}})

	gsr.RegisterBuiltin("complex128", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(complex128), nil
	}})

	gsr.RegisterBuiltin("bool", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(bool), nil
	}})

	gsr.RegisterBuiltin("string", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(string), nil
	}})

	gsr.RegisterBuiltin("time.Time", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(time.Time), nil
	}})

	// add Sexp types

	gsr.RegisterBuiltin("symbol", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &SexpSymbol{}, nil
	}})

	/* either:

	gsr.RegisterBuiltin("time.Time", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return new(time.Time), nil
	}})
	*/

	// PairRT *RegisteredType
	// RawRT *RegisteredType
	// ReflectRT *RegisteredType
	// ErrorRT *RegisteredType
	gsr.RegisterBuiltin("error", &RegisteredType{GenDefMap: false, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		var err error
		return &err, nil
	}})

	// SentinelRT *RegisteredType
	// ClosureRT *RegisteredType
}

func TypeListFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 0 {
		return SexpNull, WrongNargs
	}
	r := ListRegisteredTypes
	s := make([]Sexp, len(r))
	for i := range r {
		s[i] = &SexpStr{S: r[i]}
	}
	return env.NewSexpArray(s), nil
}

func (env *Zlisp) ImportBaseTypes() {
	for _, e := range GoStructRegistry.Builtin {
		env.AddGlobal(e.RegisteredName, e)
	}

	for _, e := range GoStructRegistry.Userdef {
		env.AddGlobal(e.RegisteredName, e)
	}
}

func compareRegisteredTypes(a *RegisteredType, bs Sexp) (int, error) {

	var b *RegisteredType
	switch bt := bs.(type) {
	case *RegisteredType:
		b = bt
	default:
		return 0, fmt.Errorf("cannot compare %T to %T", a, bs)
	}

	if a == b {
		// equal for sure
		return 0, nil
	}
	return 1, nil
}

func (gsr *GoStructRegistryType) GetOrCreatePointerType(pointedToType *RegisteredType) *RegisteredType {
	Q("pointedToType = %#v", pointedToType)
	ptrName := "*" + pointedToType.RegisteredName
	ptrRt := gsr.Lookup(ptrName)
	if ptrRt != nil {
		Q("type named '%v' already registered, reusing the pointer type", ptrName)
	} else {
		Q("registering new pointer type '%v'", ptrName)
		derivedType := reflect.PtrTo(pointedToType.TypeCache)
		ptrRt = NewRegisteredType(func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return reflect.New(derivedType), nil
		})
		ptrRt.DisplayAs = fmt.Sprintf("(* %s)", pointedToType.DisplayAs)
		ptrRt.RegisteredName = ptrName
		gsr.RegisterUserdef(ptrRt, false, ptrName)
	}
	return ptrRt
}

func (gsr *GoStructRegistryType) GetOrCreateSliceType(rt *RegisteredType) *RegisteredType {
	//sliceName := "sliceOf" + rt.RegisteredName
	sliceName := "[]" + rt.RegisteredName
	sliceRt := gsr.Lookup(sliceName)
	if sliceRt != nil {
		Q("type named '%v' already registered, re-using the type", sliceName)
	} else {
		Q("registering new slice type '%v'", sliceName)
		derivedType := reflect.SliceOf(rt.TypeCache)
		sliceRt = NewRegisteredType(func(env *Zlisp, h *SexpHash) (interface{}, error) {
			return reflect.MakeSlice(derivedType, 0, 0), nil
		})
		sliceRt.DisplayAs = fmt.Sprintf("(%s)", sliceName)
		sliceRt.RegisteredName = sliceName
		gsr.RegisterUserdef(sliceRt, false, sliceName)
	}
	return sliceRt
}

func RegisterDemoStructs() {

	gsr := &GoStructRegistry

	// demo and user defined structs
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Event{}, nil
	}}, true, "eventdemo")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Person{}, nil
	}}, true, "persondemo")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Snoopy{}, nil
	}}, true, "snoopy")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Hornet{}, nil
	}}, true, "hornet")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Hellcat{}, nil
	}}, true, "hellcat")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Weather{}, nil
	}}, true, "weather")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &Plane{}, nil
	}}, true, "plane")
	gsr.RegisterUserdef(&RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &SetOfPlanes{}, nil
	}}, true, "setOfPlanes")
}
