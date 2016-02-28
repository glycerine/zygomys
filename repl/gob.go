package zygo

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func GobEncodeFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	h, isHash := args[0].(*SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("gob argument must be a hash or defmap")
	}

	// fill the go shadow struct
	_, err := ToGoFunction(env, "togo", []Sexp{h})
	if err != nil {
		return SexpNull, fmt.Errorf("error converting object to Go struct: '%s'", err)
	}

	// serialize to gob
	var gobBytes bytes.Buffer

	enc := gob.NewEncoder(&gobBytes)
	err = enc.Encode(h.GoShadowStruct)
	if err != nil {
		return SexpNull, fmt.Errorf("gob encode error: '%s'", err)
	}

	return &SexpRaw{Val: gobBytes.Bytes()}, nil
}

func GobDecodeFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	raw, isRaw := args[0].(*SexpRaw)
	if !isRaw {
		return SexpNull, fmt.Errorf("ungob argument must be raw []byte")
	}

	rawBuf := bytes.NewBuffer(raw.Val)
	dec := gob.NewDecoder(rawBuf)
	var iface interface{}
	err := dec.Decode(iface)
	if err != nil {
		return SexpNull, fmt.Errorf("gob decode error: '%s'", err)
	}

	// TODO convert to hash
	panic("not done yet!")

	//return SexpNull, nil
}
