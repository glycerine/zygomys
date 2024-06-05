package zygo

import (
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
)

var NoAttachedGoStruct = fmt.Errorf("hash has no attach Go struct")

func HashExpression(env *Zlisp, expr Sexp) (int, error) {

	hashcode, isList, err := hashHelper(expr)
	if err != nil {
		return 0, err
	}
	if !isList {
		return hashcode, nil
	}

	// can we evaluate it?
	if env != nil {
		res, err := EvalFunction(env, "eval-hash-key", []Sexp{expr})
		if err != nil {
			return 0, fmt.Errorf("error during eval of "+
				"hash key: %s", err)
		}
		// 2nd try
		hashcode2, isList2, err := hashHelper(res)
		if err != nil {
			return 0, fmt.Errorf("evaluated key function to '%s' but could not hash type %T: %s", res.SexpString(nil), res, err)
		}
		if !isList2 {
			return hashcode2, nil
		}
		return 0, fmt.Errorf("list '%s' found where hash key needed", res.SexpString(nil))
	} // end if env == nil
	return 0, fmt.Errorf("cannot hash type %T", expr)
}

func hashHelper(expr Sexp) (hashcode int, isList bool, err error) {
	switch e := expr.(type) {
	case *SexpInt:
		return int(e.Val), false, nil
	case *SexpChar:
		return int(e.Val), false, nil
	case *SexpSymbol:
		return e.number, false, nil
	case *SexpStr:
		hasher := fnv.New32()
		_, err := hasher.Write([]byte(e.S))
		if err != nil {
			return 0, false, err
		}
		return int(hasher.Sum32()), false, nil
	case *SexpPair:
		return 0, true, nil
	case *SexpArray:
		return int(Blake2bUint64([]byte(e.SexpString(nil)))), false, nil
	}
	return 0, false, fmt.Errorf("cannot hash type %T", expr)
}

func MakeHash(args []Sexp, typename string, env *Zlisp) (*SexpHash, error) {
	//	Q("MakeHash called ")
	//	for i := range args {
	//		Q("MakeHash args[i=%v] = '%v'", i, args[i].SexpString(nil))
	//	}

	// when passed for example (hash [0]:12) we see
	// 3 args -- the colon is passed as the colon function;
	// so eliminate it as it is just an
	// extra unwanted element. This means we can never store
	// the colon function in a hash; that's okay; its
	// purpose is convenient syntax.
	args = env.EliminateColonAndCommaFromArgs(args)
	if len(args)%2 != 0 {
		return &SexpHash{Env: env},
			errors.New("hash requires even number of arguments")
	}

	var memberCount int
	var arr SexpArray
	var fld SexpArray
	var meth = []reflect.Method{}
	var field = []reflect.StructField{}
	var va reflect.Value
	num := -1
	var got reflect.Type
	var iface interface{}
	jsonMap := make(map[string]*HashFieldDet)

	factory := GoStructRegistry.Lookup(typename)
	if factory == nil {
		factory = &RegisteredType{Factory: MakeGoStructFunc(func(env *Zlisp, h *SexpHash) (interface{}, error) { return MakeHash(nil, typename, env) })}
		factory.Aliases = make(map[string]bool)
	}
	// how about UserStructDefn ? if TypeName != field/hash

	detOrder := []*HashFieldDet{}

	var zmain SexpFunction
	zmethods := make(map[string]*SexpFunction)
	var superClass *SexpHash
	var defnEnv *SexpHash

	//Q("generating SexpHash with typename: '%s'", typename)
	hash := SexpHash{
		TypeName:         typename,
		Map:              make(map[int][]*SexpPair),
		KeyOrder:         []Sexp{},
		GoStructFactory:  factory,
		NumKeys:          memberCount,
		GoMethods:        meth,
		GoMethSx:         arr,
		GoFieldSx:        fld,
		GoFields:         field,
		NumMethod:        num,
		GoType:           got,
		JsonTagMap:       jsonMap,
		GoShadowStructVa: va,
		GoShadowStruct:   iface,
		DetOrder:         detOrder,
		ZMain:            zmain,
		ZMethods:         zmethods,
		SuperClass:       superClass,
		DefnEnv:          defnEnv,
		Env:              env,
	}
	k := 0
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		err := hash.HashSet(key, val)
		if err != nil {
			return &hash, err
		}
		k++
	}

	//Q("doing factory, foundRecordType := GoStructRegistry.Registry[typename]")
	factoryShad, foundRecordType := GoStructRegistry.Registry[typename]
	if foundRecordType {
		//Q("factoryShad = '%#v' for typename='%s'\n", factoryShad, typename)
		if factoryShad.hasShadowStruct {
			//Q("\n in MakeHash: found struct associated with '%s'\n", typename)
			hash.SetGoStructFactory(factoryShad)
			//Q("\n in MakeHash: after SetGoStructFactory for typename '%s'\n", typename)
			err := hash.SetMethodList(env)
			if err != nil {
				return &SexpHash{Env: env}, fmt.Errorf("unexpected error "+
					"from hash.SetMethodList(): %s", err)
			}
		} else {
			err := factoryShad.TypeCheckRecord(&hash)
			if err != nil {
				return &SexpHash{Env: env}, err
			}
		}
	} else {
		//Q("\n in MakeHash: did not find Go struct with typename = '%s'\n", typename)
		factory.initDone = true
		factory.ReflectName = typename
		factory.DisplayAs = typename

		GoStructRegistry.RegisterUserdef(factory, false, typename)
	}

	return &hash, nil
}

func (h *SexpHash) DotPathHashGet(env *Zlisp, sym *SexpSymbol) (Sexp, error) {
	path := DotPartsRegex.FindAllString(sym.name, -1)
	//Q("in DotPathHashGet(), path = '%#v'", path)
	if len(path) == 0 {
		return SexpNull, fmt.Errorf("internal error: DotPathHashGet" +
			" path had zero length")
	}

	//	Q("\n in DotPathHashGet(), about to call nestedPathGetSet() with"+
	//		"path='%#v\n", path)
	exp, err := h.nestedPathGetSet(env, path, nil)
	if err != nil {
		return SexpNull, err
	}
	return exp, nil
}

func (hash *SexpHash) HashGet(env *Zlisp, key Sexp) (res Sexp, err error) {
	//Q("top of HashGet, key = '%v' of type %T", key.SexpString(nil), key)
	switch sym := key.(type) {
	case *SexpSymbol:
		if sym == nil {
			panic("cannot have nil symbol for key")
		}
		//P("HashGet, sym = '%v'. isDot=%v", sym.SexpString(nil), sym.isDot)
		if sym.isDot {
			return hash.DotPathHashGet(env, sym)
		}

	case *SexpArray:
		if len(sym.Val) == 1 {
			key = sym.Val[0]
		}
	}

	// this is kind of a hack
	// SexpEnd can't be created by user
	// so there is no way it would actually show up in the map
	val, err := hash.HashGetDefault(env, key, SexpEnd)

	if err != nil {
		return SexpNull, err
	}

	if val == SexpEnd {
		return SexpNull, fmt.Errorf("%s has no field '%s' [err 1]", hash.TypeName, key.SexpString(nil))
		//return SexpNull, fmt.Errorf("%s has no field '%s'", hash.UserStructDefn.Name, key.SexpString(nil))
	}
	return val, nil
}

func (hash *SexpHash) HashGetDefault(env *Zlisp, key Sexp, defaultval Sexp) (Sexp, error) {
	hashval, err := HashExpression(env, key)
	if err != nil {
		return SexpNull, err
	}
	//P("HashGetDefault, hashval='%#v', key='%s'", hashval, key.SexpString(nil))
	//for kk := range hash.Map {
	//	P("hash.Map has key '%#v'", kk)
	//}
	arr, ok := hash.Map[hashval]
	//P("arr='%#v', ok='%#v'", arr, ok)
	if !ok {
		return defaultval, nil
	}

	for _, pair := range arr {
		res, err := env.Compare(pair.Head, key)
		if err == nil && res == 0 {
			return pair.Tail, nil
		}
	}
	return defaultval, nil
}

var KeyNotSymbol = fmt.Errorf("key is not a symbol")

func (h *SexpHash) TypeCheckField(key Sexp, val Sexp) error {
	//Q("in TypeCheckField, key='%v' val='%v'", key.SexpString(nil), val.SexpString(nil))

	var keySym *SexpSymbol
	wasSym := false
	switch ks := key.(type) {
	case *SexpSymbol:
		keySym = ks
		wasSym = true
	default:
		return KeyNotSymbol
	}
	p := h.GoStructFactory
	if p == nil {
		//Q("SexpHash.TypeCheckField() sees nil GoStructFactory, bailing out.")
		return nil
	} else {
		//Q("SexpHash.TypeCheckField() sees h.GoStructFactory = '%#v'", h.GoStructFactory)
	}

	if p.UserStructDefn == nil {
		//Q("SexpHash.TypeCheckField() sees nil has.GoStructFactory.UserStructDefn, bailing out.")

		// check in the registry for this type!
		rt := GoStructRegistry.Lookup(h.TypeName)

		// was it found? If so, use it!
		if rt != nil && rt.UserStructDefn != nil {
			if rt.UserStructDefn.FieldType != nil {
				Q("")
				Q("we have a type for hash.TypeName = '%s', using it by "+
					"replacing the hash.GoStructFactory with rt", h.TypeName)
				Q("")
				Q("old: h.GoStructFactory = '%#v'", h.GoStructFactory)
				Q("")
				Q("new: rt = '%#v'", rt)
				Q("new rt.UserStructDefn.FieldType = '%#v'", rt.UserStructDefn.FieldType)
				//p.UserStructDefn = rt.UserStructDefn
				h.GoStructFactory = rt
				p = h.GoStructFactory
			}
		} else {
			return nil
		}
	}

	// type-check record updates here, if we are a record with a
	// registered type associated.
	if wasSym && h.TypeName != "hash" && h.TypeName != "field" && p != nil {
		k := keySym.name
		Q("is key '%s' defined?", k)
		declaredTyp, ok := p.UserStructDefn.FieldType[k]
		if !ok {
			return fmt.Errorf("%s has no field '%s' [err 2]", p.UserStructDefn.Name, k)
		}
		obsTyp := val.Type()
		if obsTyp == nil {
			// allow certain types to be nil, e.g. [] and nil itself
			switch a := val.(type) {
			case *SexpArray:
				if len(a.Val) == 0 {
					return nil // okay
				}
			case *SexpSentinel:
				return nil // okay
			default:
				return fmt.Errorf("%v has nil Type", val.SexpString(nil))
			}
		}

		Q("obsTyp is %T / val = %#v", obsTyp, obsTyp)
		Q("declaredTyp is %T / val = %#v", declaredTyp, declaredTyp)
		if obsTyp != declaredTyp {
			if obsTyp.RegisteredName == "[]" {
				if strings.HasPrefix(declaredTyp.RegisteredName, "[]") {
					// okay to assign empty slice to typed slice
					goto done
				}
			}
			return fmt.Errorf("field %v.%v is %v, cannot assign %v '%v'",
				p.UserStructDefn.Name,
				k,
				declaredTyp.SexpString(nil),
				obsTyp.SexpString(nil),
				val.SexpString(nil))
		}
	}
done:
	return nil
}

func (hash *SexpHash) HashSet(key Sexp, val Sexp) error {
	//vv("in HashSet, key='%v' val='%v'", key.SexpString(nil), val.SexpString(nil))

	if _, isComment := key.(*SexpComment); isComment {
		return fmt.Errorf("HashSet: key cannot be comment")
	}
	if _, isComment := val.(*SexpComment); isComment {
		return fmt.Errorf("HashSet: val cannot be comment")
	}
	if arr, isArray := key.(*SexpArray); isArray {
		na := len(arr.Val)
		if na == 1 {
			key = arr.Val[0] // let single number keys work: h[6]=10
		}
	}

	err := hash.TypeCheckField(key, val)
	if err != nil {
		if err != KeyNotSymbol {
			return err
		}
	}

	hashval, err := HashExpression(nil, key)
	if err != nil {
		return err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		hash.Map[hashval] = []*SexpPair{Cons(key, val)}
		hash.KeyOrder = append(hash.KeyOrder, key)
		hash.NumKeys++
		Q("in HashSet, added key to KeyOrder: '%v'", key)
		return nil
	}

	found := false
	for i, pair := range arr {
		res, err := hash.Env.Compare(pair.Head, key)
		if err == nil && res == 0 {
			arr[i] = Cons(key, val)
			found = true
		}
	}

	if !found {
		arr = append(arr, Cons(key, val))
		hash.KeyOrder = append(hash.KeyOrder, key)
		hash.NumKeys++
	}

	hash.Map[hashval] = arr

	return nil
}

func (hash *SexpHash) HashDelete(key Sexp) error {
	hashval, err := HashExpression(nil, key)
	if err != nil {
		return err
	}
	arr, ok := hash.Map[hashval]

	// if it doesn't exist, no need to delete it
	if !ok {
		return nil
	}

	hash.NumKeys--
	for i, pair := range arr {
		res, err := hash.Env.Compare(pair.Head, key)
		if err == nil && res == 0 {
			hash.Map[hashval] = append(arr[0:i], arr[i+1:]...)
			break
		}
	}

	return nil
}

func HashCountKeys(hash *SexpHash) int {
	var num int
	for _, arr := range hash.Map {
		num += len(arr)
	}
	if num != hash.NumKeys {
		panic(fmt.Errorf("HashCountKeys disagreement on count: num=%d, (*hash.NumKeys)=%d", num, hash.NumKeys))
	}
	return num
}

func HashIsEmpty(hash *SexpHash) bool {
	for _, arr := range hash.Map {
		if len(arr) > 0 {
			return false
		}
	}
	return true
}

func SetHashKeyOrder(hash *SexpHash, keyOrd Sexp) error {
	// truncate down to zero, then build back up correctly.
	hash.KeyOrder = hash.KeyOrder[:0]

	keys, isArr := keyOrd.(*SexpArray)
	if !isArr {
		return fmt.Errorf("must have SexpArray for keyOrd, but instead we have: %T with value='%#v'", keyOrd, keyOrd)
	}
	for _, key := range keys.Val {
		hash.KeyOrder = append(hash.KeyOrder, key)
	}

	return nil
}

func (hash *SexpHash) HashPairi(pos int) (*SexpPair, error) {
	nk := hash.NumKeys
	if pos > nk {
		return &SexpPair{}, fmt.Errorf("hpair error: pos %d is beyond our key count %d",
			pos, nk)
	}
	lenKeyOrder := len(hash.KeyOrder)
	var err error
	var key, val Sexp
	found := false
	for k := pos; k < lenKeyOrder; k++ {
		key = hash.KeyOrder[k]
		val, err = hash.HashGet(nil, key)

		if err == nil {
			found = true
			break
		}
		// what about deleted keys? just skip to the next!
	}
	if !found {
		panic(fmt.Errorf("hpair internal error: could not get element at pos %d in lenKeyOrder=%d", pos, lenKeyOrder))
	}

	return Cons(key, &SexpPair{Head: val, Tail: SexpNull}), nil
}

func GoMethodListFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(*SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw type %T/val=%#v", args[0], args[0])
	}
	if h.NumMethod != -1 {
		// use cached results
		return &h.GoMethSx, nil
	}
	// TODO: do we really need this; couldn't we check h.ShadowSet instead?
	v, err := h.GoStructFactory.Factory(env, nil)
	if v == nil {
		return SexpNull, NoAttachedGoStruct
	}
	if err != nil {
		return SexpNull, fmt.Errorf("problem during h.GoStructFactory.Factory() call: '%v'", err)
	}

	h.SetMethodList(env)
	return env.NewSexpArray(h.GoMethSx.Val), nil
}

func (h *SexpHash) SetMethodList(env *Zlisp) error {
	Q("hash.SetMethodList() called.\n")

	if !h.GoStructFactory.hasShadowStruct {
		return NoAttachedGoStruct
	}
	rs, err := h.GoStructFactory.Factory(env, nil)
	if err != nil {
		return err
	}
	if rs == nil {
		return NoAttachedGoStruct
	}
	va := reflect.ValueOf(rs)
	ty := va.Type()
	n := ty.NumMethod()

	Q("hash.SetMethodList() sees %d methods on type %v\n", n, ty)
	h.NumMethod = n
	h.GoType = ty

	sx := make([]Sexp, n)
	sl := make([]reflect.Method, n)
	for i := 0; i < n; i++ {
		sl[i] = ty.Method(i)
		sx[i] = &SexpStr{S: sl[i].Name + " " + sl[i].Type.String()}
	}
	h.GoMethSx.Val = sx
	h.GoMethods = sl

	// do the fields too

	// gotta get the struct, not a pointer to it
	e := va.Elem()
	var notAStruct = reflect.Value{}
	if e == notAStruct {
		panic(fmt.Errorf("registered GoStruct for '%s' was not a struct?!",
			h.TypeName))
	}
	tye := e.Type()
	fx := make([]Sexp, 0)
	fl := make([]reflect.StructField, 0)
	embeds := []EmbedPath{}
	json2ptr := make(map[string]*HashFieldDet)
	detOrder := make([]*HashFieldDet, 0)
	fillJsonMap(&json2ptr, &fx, &fl, embeds, tye, &detOrder)
	h.GoFieldSx.Val = fx
	h.GoFields = fl
	h.JsonTagMap = json2ptr
	h.DetOrder = detOrder
	return nil
}

const YesIamEmbeddedAbove = true

// recursively fill with embedded/anonymous types as well
func fillJsonMap(json2ptr *map[string]*HashFieldDet, fx *[]Sexp, fl *[]reflect.StructField, embedPath []EmbedPath, tye reflect.Type, detOrder *[]*HashFieldDet) {
	var suffix string
	if len(embedPath) > 0 {
		suffix = fmt.Sprintf(" embed-path<%s>", GetEmbedPath(embedPath))
	}
	m := tye.NumField()
	for i := 0; i < m; i++ {
		fld := tye.Field(i)
		*fl = append(*fl, fld)
		*fx = append(*fx, &SexpStr{S: fld.Name + " " + fld.Type.String() + suffix})
		det := &HashFieldDet{
			FieldNum:     i,
			FieldType:    fld.Type,
			StructField:  fld,
			FieldName:    fld.Name,
			FieldJsonTag: fld.Name, // fallback. changed below if json tag available.
		}
		jsonTag := fld.Tag.Get("json")
		if jsonTag != "" {
			det.FieldJsonTag = jsonTag
			(*json2ptr)[jsonTag] = det
		} else {
			(*json2ptr)[fld.Name] = det
		}
		*detOrder = append(*detOrder, det)

		// avoid cross talk between paths with a copy.
		epath := make([]EmbedPath, len(embedPath)+1)
		copy(epath, embedPath)
		epath[len(embedPath)] = EmbedPath{ChildName: fld.Name, ChildFieldNum: i}
		det.EmbedPath = epath

		if fld.Anonymous {
			// track how to get at embedded struct fields
			fillJsonMap(json2ptr, fx, fl, det.EmbedPath, fld.Type, detOrder)
		}
	}
}

func GoFieldListFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(*SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw %T/val=%v", args[0], args[0])
	}

	if !h.GoStructFactory.hasShadowStruct {
		return SexpNull, NoAttachedGoStruct
	}
	v, err := h.GoStructFactory.Factory(env, nil)
	if v == nil {
		return SexpNull, NoAttachedGoStruct
	}
	if err != nil {
		return SexpNull, fmt.Errorf("problem during h.GoStructFactory.Factory() call: '%v'", err)
	}

	return &h.GoFieldSx, nil
}

// works over hashes and arrays
func GenericHpairFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	posreq, isInt := args[1].(*SexpInt)
	if !isInt {
		return SexpNull, fmt.Errorf("hpair position request must be an integer")
	}
	pos := int(posreq.Val)

	switch seq := args[0].(type) {
	case *SexpHash:
		if pos < 0 || pos >= len(seq.KeyOrder) {
			return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
		}
		return seq.HashPairi(pos)
	case *SexpArray:
		if pos < 0 || pos >= len(seq.Val) {
			return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
		}
		return Cons(&SexpInt{Val: int64(pos)}, Cons(seq.Val[pos], SexpNull)), nil
	default:
		return SexpNull, errors.New("first argument of to hpair function must be hash, list, or array")
	}
	//return SexpNull, nil
}

func (h *SexpHash) FillHashFromShadow(env *Zlisp, src interface{}) error {
	Q("in FillHashFromShadow, with src = %#v", src)
	h.GoShadowStruct = src
	h.ShadowSet = true
	vaSrc := reflect.ValueOf(src).Elem()

	for i, det := range h.DetOrder {
		Q("\n looking at det for %s; %v-th entry in h.DetOrder\n", det.FieldJsonTag, i)
		goField := vaSrc.Field(det.FieldNum)
		val, err := fillHashHelper(goField.Interface(), 0, env, false)
		if err != nil {
			Q("got err='%s' back from fillHashhelper", err)
			return fmt.Errorf("error on GoToSexp for field '%s': '%s'",
				det.FieldJsonTag, err)
		}
		Q("got err==nil back from fillHashhelper; key=%#v, val=%#v", det.FieldJsonTag, val)
		key := env.MakeSymbol(det.FieldJsonTag)
		err = h.HashSet(key, val)
		if err != nil {
			return fmt.Errorf("error on HashSet for key '%s': '%s'", key.SexpString(nil), err)
		}
	}
	return nil
}

// translate Go -> Sexp, assuming hard *T{} struct pointers
// for all Go structs
func fillHashHelper(r interface{}, depth int, env *Zlisp, preferSym bool) (Sexp, error) {
	Q("fillHashHelper() at depth %d, decoded type is %T\n", depth, r)

	// check for one of our registered structs

	// go through the type registry upfront
	for hashName, factory := range GoStructRegistry.Registry {
		//P("fillHashHelper is trying hashName='%s'", hashName)
		st, err := factory.Factory(env, nil)
		if err != nil {
			return SexpNull, err
		}
		if reflect.ValueOf(st).Type() == reflect.ValueOf(r).Type() {
			Q("we have a registered struct match for st=%T and r=%T", st, r)
			retHash, err := MakeHash([]Sexp{}, hashName, env)
			if err != nil {
				return SexpNull, fmt.Errorf("MakeHash '%s' problem: %s",
					hashName, err)
			}

			err = retHash.FillHashFromShadow(env, r)
			if err != nil {
				return SexpNull, err
			}
			Q("retHash = %#v\n", retHash)
			return retHash, nil // or return sx?
		} else {
			Q("fillHashHelper: no match for st=%T and r=%T", st, r)
		}
	}

	Q("fillHashHelper: trying basic non-struct types for r=%T", r)

	// now handle basic non struct types:
	switch val := r.(type) {
	case string:
		Q("depth %d found string case: val = %#v\n", depth, val)
		if preferSym {
			return env.MakeSymbol(val), nil
		}
		return &SexpStr{S: val}, nil

	case int:
		Q("depth %d found int case: val = %#v\n", depth, val)
		return &SexpInt{Val: int64(val)}, nil

	case int32:
		Q("depth %d found int32 case: val = %#v\n", depth, val)
		return &SexpInt{Val: int64(val)}, nil

	case int64:
		Q("depth %d found int64 case: val = %#v\n", depth, val)
		return &SexpInt{Val: int64(val)}, nil

	case float64:
		Q("depth %d found float64 case: val = %#v\n", depth, val)
		return &SexpFloat{Val: val}, nil

	case []interface{}:
		Q("depth %d found []interface{} case: val = %#v\n", depth, val)

		slice := []Sexp{}
		for i := range val {
			sx2, err := fillHashHelper(val[i], depth+1, env, preferSym)
			if err != nil {
				return SexpNull, fmt.Errorf("error in fillHashHelper() call: '%s'", err)
			}
			slice = append(slice, sx2)
		}
		return &SexpArray{Val: slice, Env: env}, nil

	case map[string]interface{}:

		Q("depth %d found map[string]interface case: val = %#v\n", depth, val)
		sortedMapKey, sortedMapVal := makeSortedSlicesFromMap(val)

		pairs := make([]Sexp, 0)

		typeName := "hash"
		var keyOrd Sexp
		foundzKeyOrder := false
		for i := range sortedMapKey {
			// special field storing the name of our record (defmap) type.
			Q("\n i=%d sortedMapVal type %T, value=%v\n", i, sortedMapVal[i], sortedMapVal[i])
			Q("\n i=%d sortedMapKey type %T, value=%v\n", i, sortedMapKey[i], sortedMapKey[i])
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
		return hash, nil

	case []byte:
		Q("depth %d found []byte case: val = %#v\n", depth, val)

		return &SexpRaw{Val: val}, nil

	case nil:
		return SexpNull, nil

	case bool:
		return &SexpBool{Val: val}, nil

	default:
		Q("unknown type in type switch, val = %#v.  type = %T.\n", val, val)
	}

	return SexpNull, nil
}

func (h *SexpHash) nestedPathGetSet(env *Zlisp, dotpaths []string, setVal *Sexp) (Sexp, error) {

	if len(dotpaths) == 0 {
		return SexpNull, fmt.Errorf("internal error: in nestedPathGetSet() dotpaths" +
			" had zero length")
	}

	var ret Sexp = SexpNull
	var err error
	askh := h
	lenpath := len(dotpaths)
	//Q("\n in nestedPathGetSet, dotpaths=%#v\n", dotpaths)
	for i := range dotpaths {
		if setVal != nil && i == lenpath-1 {
			// assign now
			err = askh.HashSet(env.MakeSymbol(dotpaths[i][1:]), *setVal)
			//P("\n i=%v in nestedPathGetSet, dotpaths[i][1:]='%v' call to "+
			//	"HashSet returned err = '%s'\n", i, dotpaths[i][1:], err)
			return *setVal, err
		}
		ret, err = askh.HashGet(env, env.MakeSymbol(dotpaths[i][1:]))
		//P("\n i=%v in nestedPathGet, dotpaths[i][1:]='%v' call to "+
		//	"HashGet returned '%s'\n", i, dotpaths[i][1:], ret.SexpString(nil))
		if err != nil {
			return SexpNull, err
		}
		if i == lenpath-1 {
			return ret, nil
		}
		// invar: i < lenpath-1, so go deeper
		switch x := ret.(type) {
		case *SexpHash:
			//P("\n found hash in h2 at i=%d, looping to next i\n", i)
			askh = x
		case *Stack:
			return x.nestedPathGetSet(env, dotpaths[1:], setVal)
			//		case *SexpReflect:
			//			// at least allow reading, if we can.
			//			P("hashutils DEBUG! SexpReflect value x is type: '%v', '%T'", x.Val.Type(), x.Val.Interface())
			//			return SexpNull, fmt.Errorf("not a record: cannot get field '%s'"+
			//				" out of type %T)", dotpaths[i+1][1:], x)
		default:
			return SexpNull, fmt.Errorf("not a record: cannot get field '%s'"+
				" out of type %T)", dotpaths[i+1][1:], x)

		}

	}
	return ret, err
}

type ShortNamer interface {
	ShortName() string
}

func (hash *SexpHash) ShortName() string {
	return hash.TypeName
}

func (hash *SexpHash) SexpString(ps *PrintState) string {
	indInner := ""
	indent := ps.GetIndent()
	innerPs := ps.AddIndent(4) // generates a fresh new PrintState
	inner := indent + 4
	prettyEnd := ""
	origIndInner := ""
	if hash.Env.Pretty {
		prettyEnd = "\n"
		indInner = strings.Repeat(" ", inner)
		origIndInner = strings.Repeat(" ", indent)
	}
	str := " (" + hash.TypeName + " " + prettyEnd

	displayHashInCurly := false
	comma := ""
	asJSON := false
	if hash.TypeName == "hash" {
		displayHashInCurly = true
		str = "{" + prettyEnd
		if ps != nil && ps.PrintJSON {
			comma = "," // be valid JSON
			asJSON = true
		}
	}

	lastKey := hash.NumKeys
	onKey := 0
	for _, key := range hash.KeyOrder {
		val, err := hash.HashGet(hash.Env, key)
		if err == nil {
			onKey++
			switch s := key.(type) {
			case *SexpStr:
				str += indInner + `"` + s.S + `":`
			case *SexpSymbol:
				if asJSON {
					str += indInner + `"` + s.name + `":`
				} else {
					str += indInner + s.name + ":"
				}
			default:
				str += indInner + key.SexpString(innerPs) + ":"
			}
			comma2 := comma
			if onKey == lastKey {
				comma2 = ""
			}
			str += val.SexpString(innerPs) + comma2 + " " + prettyEnd

		} else {
			// ignore deleted keys
			// don't panic(err)
		}
	}
	if displayHashInCurly {
		if len(hash.Map) > 0 {
			return str[:len(str)-1] + prettyEnd + origIndInner + "}"
		}
		return str + prettyEnd + origIndInner + "}"
	}
	if len(hash.Map) > 0 {
		return str[:len(str)-1] + ")" + prettyEnd
	}
	return str + ")" + prettyEnd
}

func (r *SexpHash) Type() *RegisteredType {
	return GoStructRegistry.Registry[r.TypeName]
}

func compareHash(a *SexpHash, bs Sexp) (int, error) {

	var b *SexpHash
	switch bt := bs.(type) {
	case *SexpHash:
		b = bt
	default:
		return 0, fmt.Errorf("cannot compare %T to %T", a, bs)
	}

	if a.TypeName != b.TypeName {
		return 1, nil
	}

	return 0, nil
}

func (p *SexpHash) CopyMap() *map[int][]*SexpPair {
	cp := make(map[int][]*SexpPair)
	for k, v := range p.Map {
		cp[k] = v
	}
	return &cp
}

// CloneFrom copys all the internals of src into p, effectively
// blanking out whatever p held and replacing it with a copy of src.
func (p *SexpHash) CloneFrom(src *SexpHash) {

	p.TypeName = src.TypeName
	p.Map = *(src.CopyMap())

	p.KeyOrder = src.KeyOrder
	p.GoStructFactory = src.GoStructFactory
	p.NumKeys = src.NumKeys
	p.GoMethods = src.GoMethods
	p.GoFields = src.GoFields
	p.GoMethSx = src.GoMethSx
	p.GoFieldSx = src.GoFieldSx
	p.GoType = src.GoType
	p.NumMethod = src.NumMethod
	p.GoShadowStruct = src.GoShadowStruct
	p.GoShadowStructVa = src.GoShadowStructVa
	p.ShadowSet = src.ShadowSet

	// json tag name -> pointers to example values, as factories for SexpToGoStructs()
	p.JsonTagMap = make(map[string]*HashFieldDet)
	for k, v := range src.JsonTagMap {
		p.JsonTagMap[k] = v
	}
	p.DetOrder = src.DetOrder

	// for using these as a scoping model
	p.DefnEnv = src.DefnEnv
	p.SuperClass = src.SuperClass
	p.ZMain = src.ZMain
	p.ZMethods = make(map[string]*SexpFunction)
	for k, v := range src.ZMethods {
		p.ZMethods[k] = v
	}
	p.Env = src.Env
}

func SetPrettyPrintFlag(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg != 1 {
		return SexpNull, WrongNargs
	}
	b, isBool := args[0].(*SexpBool)
	if !isBool {
		return SexpNull, fmt.Errorf("argument to pretty must be a bool")
	}

	env.Pretty = b.Val
	return SexpNull, nil
}

// selectors for hash tables

// SexpHashSelector: reference to a symbol in a hash table.
type SexpHashSelector struct {
	Select    Sexp
	Container *SexpHash
}

func (h *SexpHash) NewSexpHashSelector(sym *SexpSymbol) *SexpHashSelector {
	return &SexpHashSelector{
		Select:    sym,
		Container: h,
	}
}

func (si *SexpHashSelector) SexpString(ps *PrintState) string {
	rhs, err := si.RHS(si.Container.Env)
	if err != nil {
		return fmt.Sprintf("SexpHashSelector error: could not get RHS: '%v'",
			err)
	}
	return fmt.Sprintf("%v /*(hashSelector %v %v)*/", rhs.SexpString(ps), si.Container.SexpString(ps), si.Select.SexpString(ps))
}

// Type returns the type of the value.
func (si *SexpHashSelector) Type() *RegisteredType {
	return GoStructRegistry.Lookup("hashSelector")
}

// RHS applies the selector to the contain and returns
// the value obtained.
func (x *SexpHashSelector) RHS(env *Zlisp) (sx Sexp, err error) {
	if env == nil {
		panic("SexpHashSelector.RSH() called with nil env")
	}
	if x.Select == nil {
		panic("cannot call RHS on hash selector with nil Select")
	}
	Q("SexpHashSelector.RHS(): x.Select is '%#v'", x.Select)
	switch t := x.Select.(type) {
	case *SexpSymbol:
		Q("SexpHashSelector.RHS(): x.Select is symbol, t = '%#v'", t)
		Q("SexpHashSelector.RHS(): x.Container is '%v'",
			x.Container.SexpString(nil))
		sx, err = x.Container.DotPathHashGet(x.Container.Env, t)
		if err != nil {
			Q("SexpHashSelector.RHS() sees err when calling"+
				" on x.Container.DotPathHashGet: with query t='%#v' err='%v'", t, err)
			return SexpNull, err
		}
	default:
		Q("SexpHashSelector.RHS() selector is not a symbol, x= '%#v'", x)
		sx, err = x.Container.HashGet(x.Container.Env, x.Select)
		if err != nil {
			Q("SexpHashSelector.RHS() sees err when calling"+
				" on x.Container.HashGet: '%v'", err)
			return SexpNull, err
		}
		//return &SexpStr{S: fmt.Sprintf("(hashidx   %s   %s)", x.Container.SexpString(nil), x.Select.SexpString(nil))}, nil
	}
	Q("SexpHashSelector) RHS() returning sx = '%v'", sx)
	return sx, nil
}

func (x *SexpHashSelector) AssignToSelection(env *Zlisp, rhs Sexp) error {
	Q("in SexpHashSelector.AssignToSelection with rhs = '%v' and container = '%v'", rhs.SexpString(nil), x.Container.SexpString(nil))
	switch sym := x.Select.(type) {
	case *SexpSymbol:
		path := DotPartsRegex.FindAllString(sym.name, -1)
		// leave dots in path, they are expected.
		_, err := x.Container.nestedPathGetSet(env, path, &rhs)
		return err
	}
	return x.Container.HashSet(x.Select, rhs)
}

// (arrayidx ar [0 1]) refers here
func HashIndexFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	Q("in HashIndexFunction, with %v args = '%#v', env=%p",
		len(args), args, env)
	for i := range args {
		Q("in HashIndexFunction, args[%v] = '%v'", i, args[i].SexpString(nil))
	}
	narg := len(args)
	if narg != 2 {
		return SexpNull, WrongNargs
	}
	tmp, err := env.ResolveDotSym([]Sexp{args[0]})
	if err != nil {
		return SexpNull, err
	}
	args[0] = tmp[0]
	Q("HashIndexFunction: past dot resolve, args[0] is now type %T/val='%v'",
		args[0], args[0].SexpString(nil))

	var hash *SexpHash
	switch ar0 := args[0].(type) {
	case *SexpHash:
		hash = ar0
	case *SexpArray:
		Q("HashIndexFunction: args[0] is an array, defering to ArrayIndexFunction")
		return ArrayIndexFunction(env, name, args)
	case Selector:
		x, err := ar0.RHS(env)
		Q("ar0.RHS() returned x = %#v", x)
		if err != nil {
			Q("HashIndexFunction: Selector error: '%v'", err)
			return SexpNull, err
		}
		switch xH := x.(type) {
		case *SexpHash:
			hash = xH
		case *SexpHashSelector:
			x, err := xH.RHS(env)
			if err != nil {
				Q("HashIndexFunction: hash retreival from "+
					"SexpHashSelector gave error: '%v'", err)
				return SexpNull, err
			}
			switch xHash2 := x.(type) {
			case *SexpHash:
				hash = xHash2
			default:
				return SexpNull, fmt.Errorf("bad (hashidx h2 index) call: h2 was a hashidx itself, but it did not resolve to an hash, instead '%s'/type %T", x.SexpString(nil), x)
			}
		case *SexpArray:
			Q("HashIndexFunction sees args[0] is Selector"+
				" that resolved to an array '%v'", xH.SexpString(nil))
			return ArrayIndexFunction(env, name, []Sexp{xH, args[1]})
		default:
			return SexpNull, fmt.Errorf("bad (hashidx h index) call: h did not resolve to a hash, instead '%s'/type %T", x.SexpString(nil), x) // failing here with x a  *SexpStr
		}
	default:
		return SexpNull, fmt.Errorf("bad (hashidx h index) call: h was not a hashmap, instead '%s'/type %T",
			args[0].SexpString(nil), args[0])
	}

	sel := args[1]
	switch x := sel.(type) {
	case *SexpSymbol:
		sel = x
		/*
			if x.isDot {
				Q("hashidx sees dot symbol: '%s', removing any prefix dot", x.name)
				if len(x.name) >= 2 && x.name[0] == '.' {
					selSym := env.MakeSymbol(x.name[1:])
					//selSym.isDot = true
					sel = selSym
				}
			}
		*/
	default:
		// okay to have SexpArray/other as selector
	}

	ret := SexpHashSelector{
		Select:    sel,
		Container: hash,
	}
	Q("HashIndexFunction: returning without error, ret.Select = '%v'", args[1].SexpString(nil))
	return &ret, nil
}

func (env *Zlisp) EliminateColonAndCommaFromArgs(args []Sexp) []Sexp {
	r := []Sexp{}
outerLoop:
	for i := range args {
		switch x := args[i].(type) {
		case *SexpComma:
			//Q("eliminating comma")
			continue outerLoop
		case *SexpFunction:
			if x.name == ":" {
				//Q("eliminating ColonFunc: args[%d] = %T/val=%#v", i, x, x)
				continue outerLoop
			}
		}
		r = append(r, args[i])
	}
	return r
}
