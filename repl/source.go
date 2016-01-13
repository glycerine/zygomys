package zygo

import (
	"fmt"
	"os"
)

func SimpleSourceFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	src, isStr := args[0].(SexpStr)
	if !isStr {
		return SexpNull, fmt.Errorf("-> error: first argument be a string")
	}

	file := string(src)
	if !FileExists(file) {
		return SexpNull, fmt.Errorf("path '%s' does not exist", file)
	}

	env2 := env.Duplicate()

	f, err := os.Open(file)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()

	err = env2.LoadFile(f)
	if err != nil {
		return SexpNull, err
	}

	_, err = env2.Run()

	return SexpNull, err
}
