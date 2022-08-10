package zygo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/glycerine/greenpack/msgp"
)

// (bsave value path) writes value as greenpack to file.
//
// (greenpack value) writes value as greenpack to SexpRaw in memory.
//
// bsave converts to binary with (togo) then saves the binary to file.
func WriteShadowGreenpackToFileFunction(name string) ZlispUserFunction {
	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {
		narg := len(args)
		if narg < 1 || narg > 2 {
			return SexpNull, WrongNargs
		}
		// check arg[0]
		var asHash *SexpHash
		switch x := args[0].(type) {
		default:
			return SexpNull, fmt.Errorf("%s error: top value must be a hash or defmap; we see '%T'", name, args[0])
		case *SexpHash:
			// okay, good
			asHash = x
		}

		switch name {
		case "bsave":
			if narg != 2 {
				return SexpNull, WrongNargs
			}

		case "greenpack":
			if narg != 1 {
				return SexpNull, WrongNargs
			}
			var buf bytes.Buffer
			_, err := toGreenpackHelper(env, asHash, &buf, "memory")
			if err != nil {
				return SexpNull, err
			}
			return &SexpRaw{Val: buf.Bytes()}, nil
		}

		// check arg[1]
		var fn string
		switch fna := args[1].(type) {
		case *SexpStr:
			fn = fna.S
		default:
			return SexpNull, fmt.Errorf("error: %s requires a string (SexpStr) path to write to as the second argument. we got type %T / value = %v", name, args[1], args[1])
		}

		// don't overwrite existing file
		if FileExists(fn) {
			return SexpNull, fmt.Errorf("error: %s refusing to write to existing file '%s'",
				name, fn)
		}

		f, err := os.Create(fn)
		if err != nil {
			return SexpNull, fmt.Errorf("error: %s sees error trying to create file '%s': '%v'", name, fn, err)
		}
		defer f.Close()

		_, err = toGreenpackHelper(env, asHash, f, fn)
		return SexpNull, err
	}
}

func toGreenpackHelper(env *Zlisp, asHash *SexpHash, f io.Writer, fn string) (Sexp, error) {

	// create shadow structs
	_, err := ToGoFunction(env, "togo", []Sexp{asHash})
	if err != nil {
		return SexpNull, fmt.Errorf("ToGo call sees error: '%v'", err)
	}

	if asHash.GoShadowStruct == nil {
		return SexpNull, fmt.Errorf("GoShadowStruct was nil, on attempt to write to '%s'", fn)
	}

	enc, ok := interface{}(asHash.GoShadowStruct).(msgp.Encodable)
	if !ok {
		return SexpNull, fmt.Errorf("error: GoShadowStruct was not greenpack Encodable -- run `go generate` or add greenpack to the source file for type '%T'. on attempt to save to '%s'", asHash.GoShadowStruct, fn)
	}
	w := msgp.NewWriter(f)
	err = msgp.Encode(w, enc)
	if err != nil {
		return SexpNull, fmt.Errorf("error: greenpack encoding to file '%s' of type '%T' sees error '%v'", fn, asHash.GoShadowStruct, err)
	}
	err = w.Flush()
	return SexpNull, err
}

func ReadGreenpackFromFileFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)

	if narg != 1 {
		return SexpNull, WrongNargs
	}
	var fn string
	switch fna := args[0].(type) {
	case *SexpStr:
		fn = fna.S
	default:
		return SexpNull, fmt.Errorf("%s requires a string path to read. we got type %T / value = %v", name, args[0], args[0])
	}

	if !FileExists(string(fn)) {
		return SexpNull, fmt.Errorf("file '%s' does not exist", fn)
	}
	f, err := os.Open(fn)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()
	by, err := ioutil.ReadAll(f)
	if err != nil {
		return SexpNull, err
	}
	return MsgpackToSexp(by, env)
}
