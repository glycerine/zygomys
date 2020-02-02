package zygo

import (
	"bytes"
	"fmt"
	"github.com/shurcooL/go-goon"
	"github.com/ugorji/go/codec"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type TypeCheckable interface {
	TypeCheck() error
}

/*
 Conversion map

 Go map[string]interface{}  <--(1)--> lisp
   ^                                  ^ |
   |                                 /  |
  (2)   ------------ (4) -----------/  (5)
   |   /                                |
   V  V                                 V
 msgpack <--(3)--> go struct, strongly typed

(1) we provide these herein; see jsonmsgp_test.go too.
     (a) SexpToGo()
     (b) GoToSexp()
(2) provided by ugorji/go/codec; see examples also herein
     (a) MsgpackToGo() / JsonToGo()
     (b) GoToMsgpack() / GoToJson()
(3) provided by tinylib/msgp, and by ugorji/go/codec
     by using pre-compiled or just decoding into an instance
     of the struct.
(4) see herein
     (a) SexpToMsgpack() and SexpToJson()
     (b) MsgpackToSexp(); uses (4) = (2) + (1)
(5) The SexpToGoStructs() and ToGoFunction() in this
    file provide the capability of marshaling an
    s-expression to a Go-struct that has been
    registered to be associated with a named
    hash map using (defmap). See repl/gotypereg.go
    to add your Go-struct constructor. From
    the prompt, the (togo) function instantiates
    a 'shadow' Go-struct whose data matches
    that configured in the record.
*/
func JsonFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch name {
	case "json":
		str := SexpToJson(args[0])
		return &SexpRaw{Val: []byte(str)}, nil
	case "unjson":
		raw, isRaw := args[0].(*SexpRaw)
		if !isRaw {
			return SexpNull, fmt.Errorf("unjson error: SexpRaw required, but we got %T instead.", args[0])
		}
		return JsonToSexp([]byte(raw.Val), env)
	case "msgpack":
		by, _ := SexpToMsgpack(args[0])
		return &SexpRaw{Val: []byte(by)}, nil
	case "unmsgpack":
		raw, isRaw := args[0].(*SexpRaw)
		if !isRaw {
			return SexpNull, fmt.Errorf("unmsgpack error: SexpRaw required, but we got %T instead.", args[0])
		}
		return MsgpackToSexp([]byte(raw.Val), env)
	default:
		return SexpNull, fmt.Errorf("JsonFunction error: unrecognized function name: '%s'", name)
	}
}

// json -> sexp. env is needed to handle symbols correctly
func JsonToSexp(json []byte, env *Zlisp) (Sexp, error) {
	iface, err := JsonToGo(json)
	if err != nil {
		return nil, err
	}
	return GoToSexp(iface, env)
}

// sexp -> json
func SexpToJson(exp Sexp) string {
	switch e := exp.(type) {
	case *SexpHash:
		return e.jsonHashHelper()
	case *SexpArray:
		return e.jsonArrayHelper()
	case *SexpSymbol:
		return `"` + e.name + `"`
	default:
		return exp.SexpString(nil)
	}
}

func (hash *SexpHash) jsonHashHelper() string {
	str := fmt.Sprintf(`{"Atype":"%s", `, hash.TypeName)

	ko := []string{}
	n := len(hash.KeyOrder)
	if n == 0 {
		return str[:len(str)-2] + "}"
	}

	for _, key := range hash.KeyOrder {
		keyst := key.SexpString(nil)
		ko = append(ko, keyst)
		val, err := hash.HashGet(nil, key)
		if err == nil {
			str += `"` + keyst + `":`
			str += string(SexpToJson(val)) + `, `
		} else {
			panic(err)
		}
	}

	str += `"zKeyOrder":[`
	for _, key := range ko {
		str += `"` + key + `", `
	}
	if n > 0 {
		str = str[:len(str)-2]
	}
	str += "]}"

	VPrintf("\n\n final ToJson() str = '%s'\n", str)
	return str
}

func (arr *SexpArray) jsonArrayHelper() string {
	if len(arr.Val) == 0 {
		return "[]"
	}

	str := "[" + SexpToJson(arr.Val[0])
	for _, sexp := range arr.Val[1:] {
		str += ", " + SexpToJson(sexp)
	}
	return str + "]"
}

type msgpackHelper struct {
	initialized bool
	mh          codec.MsgpackHandle
	jh          codec.JsonHandle
}

func (m *msgpackHelper) init() {
	if m.initialized {
		return
	}

	m.mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	// configure extensions
	// e.g. for msgpack, define functions and enable Time support for tag 1
	//does this make a differenece? m.mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	m.mh.RawToString = true
	m.mh.WriteExt = true
	m.mh.SignedInteger = true
	m.mh.Canonical = true // sort maps before writing them

	// JSON
	m.jh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	m.jh.SignedInteger = true
	m.jh.Canonical = true // sort maps before writing them

	m.initialized = true
}

var msgpHelper msgpackHelper

func init() {
	msgpHelper.init()
}

// translate to sexp -> json -> go -> msgpack
// returns both the msgpack []bytes and the go intermediary
func SexpToMsgpack(exp Sexp) ([]byte, interface{}) {

	json := []byte(SexpToJson(exp))
	iface, err := JsonToGo(json)
	panicOn(err)
	by, err := GoToMsgpack(iface)
	panicOn(err)
	return by, iface
}

// json -> go
func JsonToGo(json []byte) (interface{}, error) {
	var iface interface{}

	decoder := codec.NewDecoderBytes(json, &msgpHelper.jh)
	err := decoder.Decode(&iface)
	if err != nil {
		panic(err)
	}
	VPrintf("\n decoded type : %T\n", iface)
	VPrintf("\n decoded value: %#v\n", iface)
	return iface, nil
}

func GoToMsgpack(iface interface{}) ([]byte, error) {
	var w bytes.Buffer
	enc := codec.NewEncoder(&w, &msgpHelper.mh)
	err := enc.Encode(&iface)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// go -> json
func GoToJson(iface interface{}) []byte {
	var w bytes.Buffer
	encoder := codec.NewEncoder(&w, &msgpHelper.jh)
	err := encoder.Encode(&iface)
	if err != nil {
		panic(err)
	}
	return w.Bytes()
}

// msgpack -> sexp
func MsgpackToSexp(msgp []byte, env *Zlisp) (Sexp, error) {
	iface, err := MsgpackToGo(msgp)
	if err != nil {
		return nil, fmt.Errorf("MsgpackToSexp failed at MsgpackToGo step: '%s", err)
	}
	sexp, err := GoToSexp(iface, env)
	if err != nil {
		return nil, fmt.Errorf("MsgpackToSexp failed at GoToSexp step: '%s", err)
	}
	return sexp, nil
}

// msgpack -> go
func MsgpackToGo(msgp []byte) (interface{}, error) {

	var iface interface{}
	dec := codec.NewDecoderBytes(msgp, &msgpHelper.mh)
	err := dec.Decode(&iface)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("\n decoded type : %T\n", iface)
	//fmt.Printf("\n decoded value: %#v\n", iface)
	return iface, nil
}

// convert iface, which will typically be map[string]interface{},
// into an s-expression
func GoToSexp(iface interface{}, env *Zlisp) (Sexp, error) {
	return decodeGoToSexpHelper(iface, 0, env, false), nil
}

func decodeGoToSexpHelper(r interface{}, depth int, env *Zlisp, preferSym bool) (s Sexp) {

	VPrintf("decodeHelper() at depth %d, decoded type is %T\n", depth, r)
	switch val := r.(type) {
	case string:
		//VPrintf("depth %d found string case: val = %#v\n", depth, val)
		if preferSym {
			return env.MakeSymbol(val)
		}
		return &SexpStr{S: val}

	case int:
		VPrintf("depth %d found int case: val = %#v\n", depth, val)
		return &SexpInt{Val: int64(val)}

	case int32:
		VPrintf("depth %d found int32 case: val = %#v\n", depth, val)
		return &SexpInt{Val: int64(val)}

	case int64:
		VPrintf("depth %d found int64 case: val = %#v\n", depth, val)
		return &SexpInt{Val: val}

	case float64:
		VPrintf("depth %d found float64 case: val = %#v\n", depth, val)
		return &SexpFloat{Val: val}

	case []interface{}:
		VPrintf("depth %d found []interface{} case: val = %#v\n", depth, val)

		slice := []Sexp{}
		for i := range val {
			slice = append(slice, decodeGoToSexpHelper(val[i], depth+1, env, preferSym))
		}
		return &SexpArray{Val: slice, Env: env}

	case map[string]interface{}:

		VPrintf("depth %d found map[string]interface case: val = %#v\n", depth, val)
		sortedMapKey, sortedMapVal := makeSortedSlicesFromMap(val)

		pairs := make([]Sexp, 0)

		typeName := "hash"
		var keyOrd Sexp
		foundzKeyOrder := false
		for i := range sortedMapKey {
			// special field storing the name of our record (defmap) type.
			VPrintf("\n i=%d sortedMapVal type %T, value=%v\n", i, sortedMapVal[i], sortedMapVal[i])
			VPrintf("\n i=%d sortedMapKey type %T, value=%v\n", i, sortedMapKey[i], sortedMapKey[i])
			if sortedMapKey[i] == "zKeyOrder" {
				keyOrd = decodeGoToSexpHelper(sortedMapVal[i], depth+1, env, true)
				foundzKeyOrder = true
			} else if sortedMapKey[i] == "Atype" {
				tn, isString := sortedMapVal[i].(string)
				if isString {
					typeName = string(tn)
				}
			} else {
				sym := env.MakeSymbol(sortedMapKey[i])
				pairs = append(pairs, sym)
				ele := decodeGoToSexpHelper(sortedMapVal[i], depth+1, env, preferSym)
				pairs = append(pairs, ele)
			}
		}
		hash, err := MakeHash(pairs, typeName, env)
		if foundzKeyOrder {
			err = SetHashKeyOrder(hash, keyOrd)
			panicOn(err)
		}
		panicOn(err)
		return hash

	case []byte:
		VPrintf("depth %d found []byte case: val = %#v\n", depth, val)

		return &SexpRaw{Val: val}

	case nil:
		return SexpNull

	case bool:
		return &SexpBool{Val: val}

	case *SexpReflect:
		return decodeGoToSexpHelper(val.Val.Interface(), depth+1, env, preferSym)

	case time.Time:
		return &SexpTime{Tm: val}

	default:
		// do we have a struct for it?
		nm := fmt.Sprintf("%T", val)
		rt := GoStructRegistry.Lookup(nm)
		if rt == nil {
			fmt.Printf("unknown type '%s' in type switch, val = %#v.  type = %T.\n", nm, val, val)
		} else {
			fmt.Printf("TODO: known struct '%s' in GoToSexp(), val = %#v.  type = %T. TODO: make a record for it.\n", nm, val, val)
		}
		return SexpNull
	}

	return s
}

//msgp:ignore mapsorter KiSlice

type mapsorter struct {
	key   string
	iface interface{}
}

type KiSlice []*mapsorter

func (a KiSlice) Len() int           { return len(a) }
func (a KiSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a KiSlice) Less(i, j int) bool { return a[i].key < a[j].key }

func makeSortedSlicesFromMap(m map[string]interface{}) ([]string, []interface{}) {
	key := make([]string, len(m))
	val := make([]interface{}, len(m))
	so := make(KiSlice, 0)
	for k, i := range m {
		so = append(so, &mapsorter{key: k, iface: i})
	}
	sort.Sort(so)
	for i := range so {
		key[i] = so[i].key
		val[i] = so[i].iface
	}
	return key, val
}

// translate an Sexpr to a go value that doesn't
// depend on any Sexp/Zlisp types. Zlisp maps
// will get turned into map[string]interface{}.
// This is mostly just an exercise in type conversion.
//
// on first entry, dedup can be nil. We use it to write the
//  same pointer for a SexpHash used in more than one place.
//
func SexpToGo(sexp Sexp, env *Zlisp, dedup map[*SexpHash]interface{}) (result interface{}) {

	cacheHit := false
	if dedup == nil {
		dedup = make(map[*SexpHash]interface{})
	}

	defer func() {
		recov := recover()
		if !cacheHit && recov == nil {
			asHash, ok := sexp.(*SexpHash)
			if ok {
				// cache it. we might be overwriting with
				// ourselves, but faster to just write again
				// than to read and compare then write.
				dedup[asHash] = result
				//P("dedup caching in SexpToGo for hash %p / name='%s'", result, asHash.TypeName)
			}
		}
		if recov != nil {
			panic(recov)
		} else {
			tc, ok := result.(TypeCheckable)
			if ok {
				err := tc.TypeCheck()
				if err != nil {
					panic(fmt.Errorf("TypeCheck() error in zygo.SexpToGo for '%T': '%v'", result, err))
				}
			}
		}
	}()

	switch e := sexp.(type) {
	case *SexpRaw:
		return []byte(e.Val)
	case *SexpArray:
		//P("*SexpArray decoding! e.Val='%#v'", e.Val)
		ar := make([]interface{}, len(e.Val))
		for i, ele := range e.Val {
			ar[i] = SexpToGo(ele, env, dedup)
		}
		return ar
	case *SexpInt:
		// ugorji msgpack will give us int64 not int,
		// so match that to make the decodings comparable.
		return int64(e.Val)
	case *SexpStr:
		return e.S
	case *SexpChar:
		return rune(e.Val)
	case *SexpFloat:
		return float64(e.Val)
	case *SexpHash:

		// check dedup cache to see if we already generated a Go
		// struct for this *SexpHash.
		if alreadyGo, already := dedup[e]; already {
			//P("SexpToGo dedup cache HIT! woot! alreadyGo = '%v' for src.TypeName='%v'", alreadyGo, e.TypeName)
			cacheHit = true
			return alreadyGo
		}

		m := make(map[string]interface{})
		for _, arr := range e.Map {
			for _, pair := range arr {
				key := SexpToGo(pair.Head, env, dedup)
				val := SexpToGo(pair.Tail, env, dedup)
				keyString, isStringKey := key.(string)
				if !isStringKey {
					panic(fmt.Errorf("key '%v' should have been a string, but was not.", key))
				}
				m[keyString] = val
			}
		}
		m["Atype"] = e.TypeName
		ko := make([]interface{}, 0)
		for _, k := range e.KeyOrder {
			ko = append(ko, SexpToGo(k, env, dedup))
		}
		m["zKeyOrder"] = ko
		return m
	case *SexpPair:
		// no conversion
		return e
	case *SexpSymbol:
		return e.name
	case *SexpFunction:
		// no conversion done
		return e
	case *SexpSentinel:
		// no conversion done
		return e
	case *SexpBool:
		return e.Val
	default:
		fmt.Printf("\n error: unknown type: %T in '%#v'\n", e, e)
	}
	return nil
}

func ToGoFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch asHash := args[0].(type) {
	default:
		return SexpNull, fmt.Errorf("ToGoFunction (togo) error: value must be a hash or defmap; we see '%T'", args[0])
	case *SexpHash:
		tn := asHash.TypeName
		//P("ToGo: SexpHash for tn='%s', shadowSet='%v'", tn, asHash.ShadowSet)

		var err error
		var newStruct interface{}
		if asHash.ShadowSet && asHash.GoShadowStructVa.Kind() != reflect.Invalid {
			//P("ToGo: tn '%s' already has GoShadowStruct, not making a new one", tn)

			// don't return early, because we may have updates after changes
			// from the sexp hashtable side, so just set newStruct to the old
			// value and then let SexpToGoStructs() happen again.
			//return &SexpStr{S: fmt.Sprintf("%#v", asHash.GoShadowStruct)}, nil

			newStruct = asHash.GoShadowStruct
		} else {
			//P("ToGo: tn '%s' does not have GoShadowStruct set, making a new one", tn)

			factory, hasMaker := GoStructRegistry.Registry[tn]
			if !hasMaker {
				return SexpNull, fmt.Errorf("type '%s' not registered in GoStructRegistry", tn)
			}
			newStruct, err = factory.Factory(env, asHash)
			if err != nil {
				return SexpNull, err
			}
		}

		_, err = SexpToGoStructs(asHash, newStruct, env, nil)
		if err != nil {
			return SexpNull, err
		}

		// give new go struct a chance to boot up.
		if env.booter != nil {
			env.booter(newStruct)
		}
		asHash.GoShadowStruct = newStruct
		asHash.GoShadowStructVa = reflect.ValueOf(newStruct)
		asHash.ShadowSet = true
		return &SexpStr{S: fmt.Sprintf("%#v", newStruct)}, nil
	}

}

func FromGoFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	var sr *SexpReflect
	switch x := args[0].(type) {
	case *SexpReflect:
		sr = x
		return GoToSexp(x, env)
	case *SexpArraySelector:
		y, err := x.RHS(env)
		if err != nil {
			return SexpNull, err
		}
		switch z := y.(type) {
		case *SexpReflect:
			sr = z
		default:
			return SexpNull, fmt.Errorf("%s error: only works on *SexpReflect types. We saw %T inside an array selector", name, y)
		}
	default:
		return SexpNull, fmt.Errorf("%s error: only works on *SexpReflect types. We saw %T", name, args[0])
	}
	return GoToSexp(sr.Val.Interface(), env)

	return SexpNull, nil
}

func GoonDumpFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	fmt.Printf("\n")
	goon.Dump(args[0])
	return SexpNull, nil
}

// try to convert to registered go structs if possible,
// filling in the structure of target (should be a pointer).
func SexpToGoStructs(
	sexp Sexp,
	target interface{},
	env *Zlisp,
	dedup map[*SexpHash]interface{},

) (result interface{}, err error) {
	Q("top of SexpToGoStructs")
	cacheHit := false
	if dedup == nil {
		dedup = make(map[*SexpHash]interface{})
	}
	var recordKey string

	defer func() {
		_ = cacheHit
		recov := recover()
		if !cacheHit && err == nil && recov == nil {
			asHash, ok := sexp.(*SexpHash)
			if ok {
				// cache it. we might be overwriting with
				// ourselves, but faster to just write again
				// than to read and compare then write.
				dedup[asHash] = result
				//P("dedup caching in SexpToGoStructs '%v' for hash name='%s'", result, asHash.TypeName)
			}
		}
		if recov != nil {
			panic(fmt.Errorf("last recordKey field name was '%s'. Caught panic: '%v'", recordKey, recov))
		} else {
			tc, ok := result.(TypeCheckable)
			if ok {
				err := tc.TypeCheck()
				if err != nil {
					panic(fmt.Errorf("TypeCheck() error in zygo.SexpToGoStructs for '%T': '%v'", result, err))
				}
			}
		}
	}()

	Q(" 88888 entering SexpToGoStructs() with sexp=%#v and target=%#v of type %s", sexp, target, reflect.ValueOf(target).Type())
	defer func() {
		Q(" 99999 leaving SexpToGoStructs() with sexp='%#v' and target=%#v", sexp, target)
	}()

	targetIsSinglePtr := IsExactlySinglePointer(target)
	targetIsDoublePtr := IsExactlyDoublePointer(target)

	if !targetIsSinglePtr && !targetIsDoublePtr {
		Q("is not exactly single or double  pointer!!")
		panic(fmt.Errorf("SexpToGoStructs() got bad target: was not *T single level pointer, but rather %s / %T", reflect.ValueOf(target).Type(), target))
	}

	// target is a pointer to our payload.
	// targVa is a pointer to that same payload.
	targVa := reflect.ValueOf(target)
	targTyp := targVa.Type()
	targKind := targVa.Kind()
	targElemTyp := targTyp.Elem()
	targElemKind := targElemTyp.Kind()

	Q(" targVa is '%#v'", targVa)

	if targKind != reflect.Ptr {
		//		panic(fmt.Errorf("SexpToGoStructs got non-pointer type! was type %T/val=%#v.  targKind=%#v targTyp=%#v targVa=%#v", target, target, targKind, targTyp, targVa))
	}

	switch src := sexp.(type) {
	case *SexpRaw:
		targVa.Elem().Set(reflect.ValueOf([]byte(src.Val)))
	case *SexpArray:
		//Q(" starting 5555555555 on SexpArray")
		if targElemKind != reflect.Array && targElemKind != reflect.Slice {
			panic(fmt.Errorf("tried to translate from SexpArray into non-array/type: %v", targKind))
		}
		// allocate the slice
		n := len(src.Val)
		slc := reflect.MakeSlice(targElemTyp, 0, n)
		//P(" slc starts out as %v/type = %T", slc, slc.Interface())
		// if targ is *[]int, then targElem is []int, targElem.Elem() is int.
		eTyp := targElemTyp.Elem()
		for i, ele := range src.Val {
			_ = i
			goElem := reflect.New(eTyp) // returns pointer to new value
			//P(" goElem = %#v before filling i=%d", goElem, i)
			if _, err := SexpToGoStructs(ele, goElem.Interface(), env, dedup); err != nil {
				return nil, err
			}
			//P(" goElem = %#v after filling i=%d", goElem, i)
			//P(" goElem.Elem() = %#v after filling i=%d", goElem.Elem(), i)
			slc = reflect.Append(slc, goElem.Elem())
			//P(" slc after i=%d is now %v", i, slc)
		}
		targVa.Elem().Set(slc)
		//P(" targVa is now %v", targVa)

	case *SexpInt:
		// ugorji msgpack will give us int64 not int,
		// so match that to make the decodings comparable.
		//P("*SexpInt code src.Val='%#v'.. targVa.Elem()='%#v'/Type: %T", src.Val, targVa.Elem().Interface(), targVa.Elem().Interface())
		switch targVa.Elem().Interface().(type) {
		case float64:
			targVa.Elem().SetFloat(float64(src.Val))
		case int64:
			targVa.Elem().SetInt(int64(src.Val))
		default:
			targVa.Elem().SetInt(int64(src.Val))
		}
	case *SexpStr:
		targVa.Elem().SetString(src.S)
	case *SexpChar:
		targVa.Elem().Set(reflect.ValueOf(rune(src.Val)))
	case *SexpFloat:
		switch targVa.Elem().Interface().(type) {
		case int64:
			targVa.Elem().SetInt(int64(src.Val))
		case float64:
			targVa.Elem().SetFloat(float64(src.Val))
		default:
			targVa.Elem().SetFloat(float64(src.Val))
		}
	case *SexpHash:
		Q(" ==== found SexpHash")
		// check dedup cache to see if we already generated a Go
		// struct for this *SexpHash.
		if alreadyGoStruct, already := dedup[src]; already {
			Q("SexpToGoStructs dedup cache HIT! woot! alreadyGoStruct = '%v' for src.TypeName='%v'", alreadyGoStruct, src.TypeName)
			// already did it. Return alreadyGoStruct.
			cacheHit = true
			vo := reflect.ValueOf(alreadyGoStruct).Elem()
			targVa.Elem().Set(vo)

			return target, nil
		}

		tn := src.TypeName
		Q("tn='%s', target.(type) == %T", tn, target)
		if tn == "hash" {
			//  not done with 'hash' translation to Go, targTyp.Elem().Kind()='map', targTyp.Elem()='map[string]float64'
			//P(fmt.Sprintf("not done with 'hash' translation to Go, targTyp.Elem().Kind()='%v', targTyp.Elem()='%v'", targTyp.Elem().Kind(), targTyp.Elem()))
			switch target.(type) {
			case *map[string]string:
				m := make(map[string]string)
				for _, arr := range src.Map {
					for _, pair := range arr {
						key := SexpToGo(pair.Head, env, dedup)
						val := SexpToGo(pair.Tail, env, dedup)
						keys, isstr := key.(string)
						if !isstr {
							panic(fmt.Errorf("key '%v' should have been an string, but was not.", key))
						}
						vals, isstr := val.(string)
						if !isstr {
							panic(fmt.Errorf("val '%v' should have been an string, but was not.", val))
						}
						m[keys] = vals
					}
				}
				targVa.Elem().Set(reflect.ValueOf(m))
				return target, nil

			case *map[int64]float64:
				//P("target is a map[int64]float64")

				m := make(map[int64]float64)
				for _, arr := range src.Map {
					for _, pair := range arr {
						key := SexpToGo(pair.Head, env, dedup)
						val := SexpToGo(pair.Tail, env, dedup)
						keyint64, isint64Key := key.(int64)
						if !isint64Key {
							panic(fmt.Errorf("key '%v' should have been an int64, but was not.", key))
						}
						switch x := val.(type) {
						case float64:
							m[keyint64] = x
						case int64:
							m[keyint64] = float64(x)
						default:
							panic(fmt.Errorf("val '%v' should have been an float64, but was not.", val))
						}
					}
				}
				targVa.Elem().Set(reflect.ValueOf(m))
				return target, nil
			}
			panic("not done here yet")
			// TODO: don't try to translate into a Go struct,
			// but instead... what? just a map[string]interface{}
			//return nil, nil
		}

		switch targTyp.Elem().Kind() {
		case reflect.Interface:
			// could be an Interface like Flyer here, that contains the struct.
		case reflect.Struct:
		// typical case
		case reflect.Ptr:
			// pointer to struct we know? if we have a factory for it below
		default:

			if targTyp.Elem().Kind() == reflect.String {
				// just write the raw sexp to the string -- done to allow non-translation
				// if a string instead of a Go struct is specified. Allows us to
				// put sexpressions into a file to record data state
				// and yet lazily defer translation into Go
				// structs at a later point, if needed at all.
				str := sexp.SexpString(nil)
				//vv("note! not doing Go struct translation, instead just returning the string '%v'", str)
				targVa.Elem().Set(reflect.ValueOf(str))
				return target, nil
			}

			Q("problem! elem kind not recognized: '%#v'/type='%T'", targTyp.Elem().Kind(), targTyp.Elem().Kind())
			panic(fmt.Errorf("tried to translate from SexpHash record into non-struct/type: %v  / targType.Elem().Kind()=%v", targKind, targTyp.Elem().Kind()))
		}

		// use targVa, but check against the type in the registry for sanity/type checking.
		factory, hasMaker := GoStructRegistry.Registry[tn]
		if !hasMaker {
			panic(fmt.Errorf("type '%s' not registered in GoStructRegistry", tn))
			//return nil, fmt.Errorf("type '%s' not registered in GoStructRegistry", tn)
		}
		//P("factory = %#v  targTyp.Kind=%s", factory, targTyp.Kind())
		checkPtrStruct, err := factory.Factory(env, src)
		if err != nil {
			return nil, err
		}
		factOutputVal := reflect.ValueOf(checkPtrStruct)
		factType := factOutputVal.Type()
		if targTyp.Kind() == reflect.Ptr && targTyp.Elem().Kind() == reflect.Interface && factType.Implements(targTyp.Elem()) {
			Q(" accepting type check: %v implements %v", factType, targTyp)

			// also here we need to allocate an actual struct in place of
			// the interface

			// caller has a pointer to an interface
			// and we just want to set that interface to point to us.
			targVa.Elem().Set(factOutputVal) // tell our caller

			// now fill into this concrete type
			targVa = factOutputVal // tell the code below
			targTyp = targVa.Type()
			targKind = targVa.Kind()
			src.ShadowSet = true
			src.GoShadowStruct = checkPtrStruct
			src.GoShadowStructVa = factOutputVal

		} else if targTyp.Kind() == reflect.Ptr && targTyp.Elem() == factType {
			Q("we have a double pointer that matches the factory type! factType == targTyp.Elem(). factType=%v/%T  targTyp = %v/%T", factType, factType, targTyp, targTyp)
			Q(" targTyp.Elem() = %v", targTyp.Elem())

			targVa.Elem().Set(factOutputVal) // tell our caller

			// now fill into this concrete type
			targVa = factOutputVal // tell the code below
			targTyp = targVa.Type()
			targKind = targVa.Kind()
			src.ShadowSet = true
			src.GoShadowStruct = checkPtrStruct
			src.GoShadowStructVa = factOutputVal

		} else if factType != targTyp {
			// factType=*zygo.NestInner/*reflect.rtype  targTyp = **zygo.NestInner/*reflect.rtype

			Q("factType != targTyp. factType=%v/%T  targTyp = %v/%T", factType, factType, targTyp, targTyp)

			Q(" targTyp.Elem() = %v", targTyp.Elem())

			panic(fmt.Errorf("type checking failed compare the factor associated with SexpHash and the provided target *T: expected '%s' (associated with typename '%s' in the GoStructRegistry) but saw '%s' type in target", tn, factType, targTyp))
		}
		//maploop:
		for _, arr := range src.Map {
			for _, pair := range arr {
				recordKey = ""
				switch k := pair.Head.(type) {
				case *SexpStr:
					recordKey = k.S
				case *SexpSymbol:
					recordKey = k.name
				default:
					fmt.Printf(" skipping field '%#v' which we don't know how to lookup.", pair.Head)
					panic(fmt.Sprintf("unknown fields disallowed: we didn't recognize '%#v'", pair.Head))
					continue
				}
				// We've got to match pair.Head to
				// one of the struct fields: we'll use
				// the json tags for that. Or their
				// full exact name if they didn't have
				// a json tag.
				Q(" JsonTagMap = %#v", src.JsonTagMap)
				det, found := src.JsonTagMap[recordKey]
				if !found {
					// try once more, with uppercased version
					// of record key
					upperKey := strings.ToUpper(recordKey[:1]) + recordKey[1:]
					det, found = src.JsonTagMap[upperKey]
					if !found {
						fmt.Printf(" skipping field '%s' in this hash/which we could not find in the JsonTagMap", recordKey)
						panic(fmt.Sprintf("unkown field '%s' not allowed; could not find in the JsonTagMap. Fieldnames are case sensitive.", recordKey))
						continue
					}
				}
				Q(" ****  recordKey = '%s'\n", recordKey)
				Q(" we found in pair.Tail: %T !", pair.Tail)

				dref := targVa.Elem()
				Q(" deref = %#v / type %T", dref, dref)

				Q(" det = %#v", det)

				// fld should hold our target when
				// done recursing through any embedded structs.
				// TODO: handle embedded pointers to structs too.
				var fld reflect.Value
				Q(" we have an det.EmbedPath of '%#v'", det.EmbedPath)
				// drill down to the actual target
				fld = dref
				for i, p := range det.EmbedPath {
					Q("about to call fld.Field(%d) on fld = '%#v'/type=%T", p.ChildFieldNum, fld, fld)
					fld = fld.Field(p.ChildFieldNum)
					Q(" dropping down i=%d through EmbedPath at '%s', fld = %#v ", i, p.ChildName, fld)
				}
				Q(" fld = %#v ", fld)

				// INVAR: fld points at our target to fill
				ptrFld := fld.Addr()
				tmp, needed := unexportHelper(&ptrFld, &fld)
				if needed {
					ptrFld = *tmp
				}
				_, err := SexpToGoStructs(pair.Tail, ptrFld.Interface(), env, dedup)
				if err != nil {
					panic(err)
					//return nil, err
				}
			}
		}
	case *SexpPair:
		panic("unimplemented")
		// no conversion
		//return src
	case *SexpSymbol:
		targVa.Elem().SetString(src.name)
	case *SexpFunction:
		panic("unimplemented: *SexpFunction converstion.")
		// no conversion done
		//return src
	case *SexpSentinel:
		// set to nil
		targVa.Elem().Set(reflect.Zero(targVa.Type().Elem()))
	case *SexpTime:
		targVa.Elem().Set(reflect.ValueOf(src.Tm))
	case *SexpBool:
		targVa.Elem().Set(reflect.ValueOf(src.Val))
	default:
		fmt.Printf("\n error: unknown type: %T in '%#v'\n", src, src)
	}
	return target, nil
}

/*
if accessing unexported fields, we'll recover from
   panic: reflect.Value.Interface: cannot return value obtained from unexported field or method
and use this technique
   https://stackoverflow.com/questions/42664837/access-unexported-fields-in-golang-reflect
*/
func unexportHelper(ptrFld *reflect.Value, fld *reflect.Value) (r *reflect.Value, needed bool) {
	defer func() {
		recov := recover()
		if recov != nil {
			//P("unexportHelper recovering from '%v'", recov)
			e := reflect.NewAt(fld.Type(), unsafe.Pointer(fld.UnsafeAddr()))
			r = &e
			needed = true
		}
	}()
	// can we do this without panic?
	_ = ptrFld.Interface()
	// if no panic, return same.
	return ptrFld, false
}

// A small set of important little buildling blocks.
// These demonstrate how to use reflect.
/*
(1) Tutorial on setting structs with reflect.Set()

http://play.golang.org/p/sDmFgZmGvv

package main

import (
"fmt"
"reflect"

)

type A struct {
  S string
}

func MakeA() interface{} {
  return &A{}
}

func main() {
   a1 := MakeA()
   a2 := MakeA()
   a2.(*A).S = "two"

   // now assign a2 -> a1 using reflect.
    targVa := reflect.ValueOf(&a1).Elem()
    targVa.Set(reflect.ValueOf(a2))
    fmt.Printf("a1 = '%#v' / '%#v'\n", a1, targVa.Interface())
}
// output
// a1 = '&main.A{S:"two"}' / '&main.A{S:"two"}'


(2) Tutorial on setting fields inside a struct with reflect.Set()

http://play.golang.org/p/1k4iQKVwUD

package main

import (
    "fmt"
    "reflect"
)

type A struct {
    S string
}

func main() {
    a1 := &A{}

    three := "three"

    fld := reflect.ValueOf(&a1).Elem().Elem().FieldByName("S")

    fmt.Printf("fld = %#v\n of type %T\n", fld, fld)
    fmt.Println("settability of fld:", fld.CanSet()) // true

    // now assign to field a1.S the string "three" using reflect.

    fld.Set(reflect.ValueOf(three))

    fmt.Printf("after fld.Set(): a1 = '%#v' \n", a1)
}

// output:
fld = ""
 of type reflect.Value
settability of fld: true
after fld.Set(): a1 = '&main.A{S:"three"}'

(3) Setting struct after passing through an function call interface{} param:

package main

import (
	"fmt"
	"reflect"
)

type A struct {
	S string
}

func main() {
	a1 := &A{}
	f(&a1)
	fmt.Printf("a1 = '%#v'\n", a1)
	// a1 = '&main.A{S:"two"}' / '&main.A{S:"two"}'
}

func f(i interface{}) {
	a2 := MakeA()
	a2.(*A).S = "two"

	// now assign a2 -> a1 using reflect.
	//targVa := reflect.ValueOf(&a1).Elem()
	targVa := reflect.ValueOf(i).Elem()
	targVa.Set(reflect.ValueOf(a2))
}

(4) using a function to do the Set(), and checking
    the received interface for correct type.
    Also: Using a function to set just one sub-field.

package main

import (
	"fmt"
	"reflect"
)

type A struct {
	S string
	R string
}

func main() {
	a1 := &A{}
	overwrite_contents_of_struct(a1)
	fmt.Printf("a1 = '%#v'\n", a1)

	// output:
	// yes, is single level pointer
	// a1 = '&main.A{S:"two", R:""}'

	assignToOnlyFieldR(a1)
	fmt.Printf("after assignToOnlyFieldR(a1):  a1 = '%#v'\n", a1)

	// output:
//	yes, is single level pointer
//	a1 = '&main.A{S:"two", R:""}'
//	yes, is single level pointer
//	fld = ""
//	 of type reflect.Value
//	settability of fld: true
//	after assignToOnlyFieldR(a1):  a1 = '&main.A{S:"two", R:"R has been altered"}'

}

func assignToOnlyFieldR(i interface{}) {
	if !IsExactlySinglePointer(i) {
		panic("not single level pointer")
	}
	fmt.Printf("yes, is single level pointer\n")

	altered := "R has been altered"

	fld := reflect.ValueOf(i).Elem().FieldByName("R")

	fmt.Printf("fld = %#v\n of type %T\n", fld, fld)
	fmt.Println("settability of fld:", fld.CanSet()) // true

	// now assign to field a1.S
	fld.Set(reflect.ValueOf(altered))
}

func overwrite_contents_of_struct(i interface{}) {
	// we want i to contain an *A, or a pointer-to struct.
	// So we can reassign *ptr = A' for a different content A'.

	if !IsExactlySinglePointer(i) {
		panic("not single level pointer")
	}
	fmt.Printf("yes, is single level pointer\n")

	a2 := &A{S: "two"}

	// now assign a2 -> a1 using reflect.
	targVa := reflect.ValueOf(i).Elem()
	targVa.Set(reflect.ValueOf(a2).Elem())
}

func IsExactlySinglePointer(target interface{}) bool {

	typ := reflect.ValueOf(target).Type()
	kind := typ.Kind()
	if kind != reflect.Ptr {
		return false
	}
	typ2 := typ.Elem()
	kind2 := typ2.Kind()
	if kind2 == reflect.Ptr {
		return false // two level pointer
	}
	return true
}

*/
