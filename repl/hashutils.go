package zygo

import (
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
)

var NoAttachedGoStruct = fmt.Errorf("hash has no attach Go struct")

func HashExpression(env *Glisp, expr Sexp) (int, error) {

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
			return 0, fmt.Errorf("evaluated key function to '%s' but could not hash type %T: %s", res.SexpString(), res, err)
		}
		if !isList2 {
			return hashcode2, nil
		}
		return 0, fmt.Errorf("list '%s' found where hash key needed", res.SexpString())
	}
	return 0, fmt.Errorf("cannot hash type %T", expr)
}

func hashHelper(expr Sexp) (hashcode int, isList bool, err error) {
	switch e := expr.(type) {
	case SexpInt:
		return int(e), false, nil
	case SexpChar:
		return int(e), false, nil
	case SexpSymbol:
		return e.number, false, nil
	case SexpStr:
		hasher := fnv.New32()
		_, err := hasher.Write([]byte(e))
		if err != nil {
			return 0, false, err
		}
		return int(hasher.Sum32()), false, nil
	case SexpPair:
		return 0, true, nil
	}
	return 0, false, fmt.Errorf("cannot hash type %T", expr)
}

func MakeHash(args []Sexp, typename string, env *Glisp) (SexpHash, error) {
	if len(args)%2 != 0 {
		return SexpHash{},
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
	factory := RegistryEntry{Factory: MakeGoStructFunc(func(env *Glisp) interface{} { return nil })}
	detOrder := []*HashFieldDet{}

	var zmain SexpFunction
	zmethods := make(map[string]*SexpFunction)
	var superClass *SexpHash
	var defnEnv *SexpHash

	hash := SexpHash{
		TypeName:         &typename,
		Map:              make(map[int][]SexpPair),
		KeyOrder:         &[]Sexp{},
		GoStructFactory:  &factory,
		NumKeys:          &memberCount,
		GoMethods:        &meth,
		GoMethSx:         &arr,
		GoFieldSx:        &fld,
		GoFields:         &field,
		NumMethod:        &num,
		GoType:           &got,
		JsonTagMap:       &jsonMap,
		GoShadowStructVa: &va,
		GoShadowStruct:   &iface,
		DetOrder:         &detOrder,
		ZMain:            &zmain,
		ZMethods:         &zmethods,
		SuperClass:       &superClass,
		DefnEnv:          &defnEnv,
		env:              env,
	}
	k := 0
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		err := hash.HashSet(key, val)
		if err != nil {
			return hash, err
		}
		k++
	}

	factory, foundGoStruct := GostructRegistry[typename]
	if foundGoStruct {
		VPrintf("\n in MakeHash: found struct associated with '%s'\n", typename)
		hash.SetGoStructFactory(factory)
		err := hash.SetMethodList(env)
		if err != nil {
			return SexpHash{}, fmt.Errorf("unexpected error "+
				"from hash.SetMethodList(): %s", err)
		}
	} else {
		VPrintf("\n in MakeHash: did not find Go struct with '%s'\n", typename)
	}

	return hash, nil
}

func (hash *SexpHash) HashGet(env *Glisp, key Sexp) (Sexp, error) {
	// this is kind of a hack
	// SexpEnd can't be created by user
	// so there is no way it would actually show up in the map
	val, err := hash.HashGetDefault(env, key, SexpEnd)

	if err != nil {
		return SexpNull, err
	}

	if val == SexpEnd {
		msg := fmt.Sprintf("key %s not found", key.SexpString())
		return SexpNull, errors.New(msg)
	}
	return val, nil
}

func (hash *SexpHash) HashGetDefault(env *Glisp, key Sexp, defaultval Sexp) (Sexp, error) {
	hashval, err := HashExpression(env, key)
	if err != nil {
		return SexpNull, err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		return defaultval, nil
	}

	for _, pair := range arr {
		res, err := Compare(pair.Head, key)
		if err == nil && res == 0 {
			return pair.Tail, nil
		}
	}
	return defaultval, nil
}

func (hash *SexpHash) HashSet(key Sexp, val Sexp) error {
	hashval, err := HashExpression(nil, key)
	if err != nil {
		return err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		hash.Map[hashval] = []SexpPair{Cons(key, val)}
		*hash.KeyOrder = append(*hash.KeyOrder, key)
		(*hash.NumKeys)++
		return nil
	}

	found := false
	for i, pair := range arr {
		res, err := Compare(pair.Head, key)
		if err == nil && res == 0 {
			arr[i] = Cons(key, val)
			found = true
		}
	}

	if !found {
		arr = append(arr, Cons(key, val))
		*hash.KeyOrder = append(*hash.KeyOrder, key)
		(*hash.NumKeys)++
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

	(*hash.NumKeys)--
	for i, pair := range arr {
		res, err := Compare(pair.Head, key)
		if err == nil && res == 0 {
			hash.Map[hashval] = append(arr[0:i], arr[i+1:]...)
			break
		}
	}

	return nil
}

func HashCountKeys(hash SexpHash) int {
	var num int
	for _, arr := range hash.Map {
		num += len(arr)
	}
	if num != (*hash.NumKeys) {
		panic(fmt.Errorf("HashCountKeys disagreement on count: num=%d, (*hash.NumKeys)=%d", num, (*hash.NumKeys)))
	}
	return num
}

func HashIsEmpty(hash SexpHash) bool {
	for _, arr := range hash.Map {
		if len(arr) > 0 {
			return false
		}
	}
	return true
}

func SetHashKeyOrder(hash *SexpHash, keyOrd Sexp) error {
	// truncate down to zero, then build back up correctly.
	*(*hash).KeyOrder = (*(*hash).KeyOrder)[:0]

	keys, isArr := keyOrd.(SexpArray)
	if !isArr {
		return fmt.Errorf("must have SexpArray for keyOrd, but instead we have: %T with value='%#v'", keyOrd, keyOrd)
	}
	for _, key := range keys {
		*hash.KeyOrder = append(*hash.KeyOrder, key)
	}

	return nil
}

func (hash *SexpHash) HashPairi(pos int) (SexpPair, error) {
	nk := (*hash.NumKeys)
	if pos > nk {
		return SexpPair{}, fmt.Errorf("hpair error: pos %d is beyond our key count %d",
			pos, nk)
	}
	lenKeyOrder := len(*hash.KeyOrder)
	var err error
	var key, val Sexp
	found := false
	for k := pos; k < lenKeyOrder; k++ {
		key = (*hash.KeyOrder)[k]
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

	return Cons(key, SexpPair{Head: val, Tail: SexpNull}), nil
}

func GoMethodListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw type %T/val=%#v", args[0], args[0])
	}
	if *h.NumMethod != -1 {
		// use cached results
		return *h.GoMethSx, nil
	}
	if (*h.GoStructFactory).Factory(env) == nil {
		return SexpNull, NoAttachedGoStruct
	}

	h.SetMethodList(env)
	return SexpArray(*h.GoMethSx), nil
}

func (h *SexpHash) SetMethodList(env *Glisp) error {
	VPrintf("hash.SetMethodList() called.\n")

	rs := (*h.GoStructFactory).Factory(env)
	if rs == nil {
		return NoAttachedGoStruct
	}
	va := reflect.ValueOf(rs)
	ty := va.Type()
	n := ty.NumMethod()

	VPrintf("hash.SetMethodList() sees %d methods on type %v\n", n, ty)
	*h.NumMethod = n
	*h.GoType = ty

	sx := make([]Sexp, n)
	sl := make([]reflect.Method, n)
	for i := 0; i < n; i++ {
		sl[i] = ty.Method(i)
		sx[i] = SexpStr(sl[i].Name + " " + sl[i].Type.String())
	}
	*h.GoMethSx = sx
	*h.GoMethods = sl

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
	*h.GoFieldSx = fx
	*h.GoFields = fl
	*h.JsonTagMap = json2ptr
	*h.DetOrder = detOrder
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
		*fx = append(*fx, SexpStr(fld.Name+" "+fld.Type.String()+suffix))
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
		det.EmbedPath = append(embedPath,
			EmbedPath{ChildName: fld.Name, ChildFieldNum: i})
		if fld.Anonymous {
			// track how to get at embedded struct fields
			fillJsonMap(json2ptr, fx, fl, det.EmbedPath, fld.Type, detOrder)
		}
	}
}

func GoFieldListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw %T/val=%v", args[0], args[0])
	}
	if (*h.GoStructFactory).Factory(env) == nil {
		return SexpNull, NoAttachedGoStruct
	}
	return SexpArray(*h.GoFieldSx), nil
}

// works over hashes and arrays
func GenericHpairFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	posreq, isInt := args[1].(SexpInt)
	if !isInt {
		return SexpNull, fmt.Errorf("hpair position request must be an integer")
	}
	pos := int(posreq)

	switch seq := args[0].(type) {
	case SexpHash:
		if pos < 0 || pos >= len(*seq.KeyOrder) {
			return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
		}
		return seq.HashPairi(pos)
	case SexpArray:
		if pos < 0 || pos >= len(seq) {
			return SexpNull, fmt.Errorf("hpair position request %d out of bounds", pos)
		}
		return Cons(SexpInt(pos), Cons(seq[pos], SexpNull)), nil
	default:
		return SexpNull, errors.New("first argument of to hpair function must be hash, list, or array")
	}
	return SexpNull, nil
}

func (h *SexpHash) FillHashFromShadow(env *Glisp, src interface{}) error {
	(*h.GoShadowStruct) = src
	vaSrc := reflect.ValueOf(src).Elem()

	for _, det := range *h.DetOrder {
		VPrintf("\n looking at det for %s\n", det.FieldJsonTag)
		goField := vaSrc.Field(det.FieldNum)
		val, err := fillHashHelper(goField.Interface(), 0, env, false)
		if err != nil {
			return fmt.Errorf("error on GoToSexp for field '%s': '%s'",
				det.FieldJsonTag, err)
		}
		key := env.MakeSymbol(det.FieldJsonTag)
		err = h.HashSet(key, val)
		if err != nil {
			return fmt.Errorf("error on HashSet for key '%s': '%s'", key, err)
		}
	}
	return nil
}

// translate Go -> Sexp, assuming hard *T{} struct pointers
// for all Go structs
func fillHashHelper(r interface{}, depth int, env *Glisp, preferSym bool) (Sexp, error) {
	VPrintf("fillHashHelper() at depth %d, decoded type is %T\n", depth, r)

	// check for one of our registered structs

	// go through the type registry upfront
	for hashName, factory := range GostructRegistry {
		st := factory.Factory(env)
		if reflect.ValueOf(st).Type() == reflect.ValueOf(r).Type() {
			retHash, err := MakeHash([]Sexp{}, hashName, env)
			if err != nil {
				return SexpNull, fmt.Errorf("MakeHash '%s' problem: %s",
					hashName, err)
			}

			err = retHash.FillHashFromShadow(env, r)
			if err != nil {
				return SexpNull, err
			}
			VPrintf("retHash = %#v\n", retHash)
			return retHash, nil // or return sx?
		}
	}

	// now handle basic non struct types:
	switch val := r.(type) {
	case string:
		VPrintf("depth %d found string case: val = %#v\n", depth, val)
		if preferSym {
			return env.MakeSymbol(val), nil
		}
		return SexpStr(val), nil

	case int:
		VPrintf("depth %d found int case: val = %#v\n", depth, val)
		return SexpInt(val), nil

	case int32:
		VPrintf("depth %d found int32 case: val = %#v\n", depth, val)
		return SexpInt(val), nil

	case int64:
		VPrintf("depth %d found int64 case: val = %#v\n", depth, val)
		return SexpInt(val), nil

	case float64:
		VPrintf("depth %d found float64 case: val = %#v\n", depth, val)
		return SexpFloat(val), nil

	case []interface{}:
		VPrintf("depth %d found []interface{} case: val = %#v\n", depth, val)

		slice := []Sexp{}
		for i := range val {
			sx2, err := fillHashHelper(val[i], depth+1, env, preferSym)
			if err != nil {
				return SexpNull, fmt.Errorf("error in fillHashHelper() call: '%s'", err)
			}
			slice = append(slice, sx2)
		}
		return SexpArray(slice), nil

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
			err = SetHashKeyOrder(&hash, keyOrd)
			panicOn(err)
		}
		panicOn(err)
		return hash, nil

	case []byte:
		VPrintf("depth %d found []byte case: val = %#v\n", depth, val)

		return SexpRaw(val), nil

	case nil:
		return SexpNull, nil

	case bool:
		return SexpBool(val), nil

	default:
		VPrintf("unknown type in type switch, val = %#v.  type = %T.\n", val, val)
	}

	return SexpNull, nil
}

func (h *SexpHash) nestedPathGetSet(env *Glisp, dotpaths []string, setVal *Sexp) (Sexp, error) {

	if len(dotpaths) == 0 {
		return SexpNull, fmt.Errorf("internal error: in nestedPathGetSet() dotpaths" +
			" had zero length")
	}

	var ret Sexp = SexpNull
	var err error
	askh := h
	lenpath := len(dotpaths)
	//P("\n in nestedPathGetSet, dotpaths=%#v\n", dotpaths)
	for i := range dotpaths {
		if setVal != nil && i == lenpath-1 {
			// assign now
			err = askh.HashSet(env.MakeSymbol(dotpaths[i][1:]), *setVal)
			//P("\n i=%v in nestedPathGetSet, dotpaths[i][1:]='%v' call to "+
			//   "HashSet returned err = '%s'\n", i, dotpaths[i][1:], err)
			return *setVal, err
		}
		ret, err = askh.HashGet(env, env.MakeSymbol(dotpaths[i][1:]))
		//P("\n i=%v in nestedPathGet, dotpaths[i][1:]='%v' call to "+
		//	"HashGet returned '%s'\n", i, dotpaths[i][1:], ret.SexpString())
		if err != nil {
			return SexpNull, err
		}
		if i == lenpath-1 {
			return ret, nil
		}
		// invar: i < lenpath-1, so go deeper
		switch h2 := ret.(type) {
		case SexpHash:
			//P("\n found hash in h2 at i=%d, looping to next i\n", i)
			askh = &h2
		default:
			return SexpNull, fmt.Errorf("not a record: cannot get field '%s'"+
				" in out of type %T)", dotpaths[i+1][1:], h2)

		}

	}
	return ret, err
}
