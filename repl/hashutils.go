package zygo

import (
	"errors"
	"fmt"
	"hash/fnv"
	"reflect"
	//"github.com/shurcooL/go-goon"
)

var NoAttachedGoStruct = fmt.Errorf("hash has no attach Go struct")

var GostructRegistry = map[string]interface{}{}

// builtin known Go Structs
func init() {
	GostructRegistry["event"] = &Event{}
	GostructRegistry["person"] = &Person{}

	GostructRegistry["snoopy"] = &Snoopy{}
	GostructRegistry["hornet"] = &Hornet{}
	GostructRegistry["hellcat"] = &Hellcat{}
}

func HashExpression(expr Sexp) (int, error) {
	switch e := expr.(type) {
	case SexpInt:
		return int(e), nil
	case SexpChar:
		return int(e), nil
	case SexpSymbol:
		return e.number, nil
	case SexpStr:
		hasher := fnv.New32()
		_, err := hasher.Write([]byte(e))
		if err != nil {
			return 0, err
		}
		return int(hasher.Sum32()), nil
	}
	return 0, errors.New(fmt.Sprintf("cannot hash type %T", expr))
}

func MakeHash(args []Sexp, typename string) (SexpHash, error) {
	if len(args)%2 != 0 {
		return SexpHash{},
			errors.New("hash requires even number of arguments")
	}

	var iface interface{}
	var memberCount int
	var arr SexpArray
	var fld SexpArray
	var meth = []reflect.Method{}
	var field = []reflect.StructField{}
	num := -1
	var got reflect.Type
	hash := SexpHash{
		TypeName:  &typename,
		Map:       make(map[int][]SexpPair),
		KeyOrder:  &[]Sexp{},
		GoStruct:  &iface,
		NumKeys:   &memberCount,
		GoMethods: &meth,
		GoMethSx:  &arr,
		GoFieldSx: &fld,
		GoFields:  &field,
		NumMethod: &num,
		GoType:    &got,
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

	stct, foundGoStruct := GostructRegistry[typename]
	if foundGoStruct {
		VPrintf("\n in MakeHash: found struct associated with '%s': %T/val=%#v\n", typename, stct, stct)
		hash.SetGoStruct(stct)
		err := hash.SetMethodList()
		if err != nil {
			return SexpHash{}, fmt.Errorf("unexpected error "+
				"from hash.SetMethodList(): %s", err)
		}
	} else {
		VPrintf("\n in MakeHash: did not find Go struct with '%s'\n", typename)
	}

	return hash, nil
}

func (hash *SexpHash) HashGet(key Sexp) (Sexp, error) {
	// this is kind of a hack
	// SexpEnd can't be created by user
	// so there is no way it would actually show up in the map
	val, err := hash.HashGetDefault(key, SexpEnd)

	if err != nil {
		return SexpNull, err
	}

	if val == SexpEnd {
		msg := fmt.Sprintf("key %s not found", key.SexpString())
		return SexpNull, errors.New(msg)
	}
	return val, nil
}

func (hash *SexpHash) HashGetDefault(key Sexp, defaultval Sexp) (Sexp, error) {
	hashval, err := HashExpression(key)
	if err != nil {
		return SexpNull, err
	}
	arr, ok := hash.Map[hashval]

	if !ok {
		return defaultval, nil
	}

	for _, pair := range arr {
		res, err := Compare(pair.head, key)
		if err == nil && res == 0 {
			return pair.tail, nil
		}
	}
	return defaultval, nil
}

func (hash *SexpHash) HashSet(key Sexp, val Sexp) error {
	hashval, err := HashExpression(key)
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
		res, err := Compare(pair.head, key)
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
	hashval, err := HashExpression(key)
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
		res, err := Compare(pair.head, key)
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
		val, err = hash.HashGet(key)

		if err == nil {
			found = true
			break
		}
		// what about deleted keys? just skip to the next!
	}
	if !found {
		panic(fmt.Errorf("hpair internal error: could not get element at pos %d in lenKeyOrder=%d", pos, lenKeyOrder))
	}

	return Cons(key, SexpPair{head: val, tail: SexpNull}), nil
}

func GoMethodListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw %T/val=%v", args[0], args[0])
	}
	if *h.NumMethod != -1 {
		// use cached results
		return *h.GoMethSx, nil
	}
	rs := reflect.ValueOf(h.GoStruct)
	if rs.IsNil() {
		return SexpNull, NoAttachedGoStruct
	}

	h.SetMethodList()
	return SexpArray(*h.GoMethSx), nil
}

func (h *SexpHash) SetMethodList() error {
	VPrintf("hash.SetMethodList() called.\n")

	rs := reflect.ValueOf(*h.GoStruct)
	VPrintf("\n in SetMethodList() rs = '%#v'\n", rs)
	if rs.IsNil() {
		return NoAttachedGoStruct
	}
	ty := rs.Type()
	n := ty.NumMethod()

	VPrintf("hash.SetMethodList() sees %d methods\n", n)
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
	e := rs.Elem()
	var notAStruct = reflect.Value{}
	if e == notAStruct {
		return fmt.Errorf("registered GoStruct was not a struct?!")
	}
	tye := e.Type()
	m := tye.NumField()
	fx := make([]Sexp, m)
	fl := make([]reflect.StructField, m)
	for i := 0; i < m; i++ {
		fl[i] = tye.Field(i)
		fx[i] = SexpStr(fl[i].Name + " " + fl[i].Type.String())
	}
	*h.GoFieldSx = fx
	*h.GoFields = fl
	return nil
}

func GoFieldListFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	h, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("hash/record required, but saw %T/val=%v", args[0], args[0])
	}
	rs := reflect.ValueOf(h.GoStruct)
	if rs.IsNil() {
		return SexpNull, NoAttachedGoStruct
	}
	return SexpArray(*h.GoFieldSx), nil
}
