package glisp

import (
	"errors"
	"fmt"
	"hash/fnv"

	//"github.com/shurcooL/go-goon"
)

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

	hash := SexpHash{
		TypeName: &typename,
		Map:      make(map[int][]SexpPair),
		KeyOrder: &[]Sexp{},
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
