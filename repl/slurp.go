package zygo

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// read new-line delimited text from a file into an array (slurpf "path-to-file")
func SlurpfileFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	var fn string
	switch fna := args[0].(type) {
	case SexpStr:
		fn = string(fna)
	default:
		return SexpNull, fmt.Errorf("slurp requires a string path to read. we got type %T / value = %v", args[0], args[0])
	}

	if !FileExists(string(fn)) {
		return SexpNull, fmt.Errorf("file '%s' does not exists", fn)
	}
	f, err := os.Open(fn)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()

	a := make([]Sexp, 0)

	bufIn := bufio.NewReader(f)
	lineNum := int64(1)
	for {
		lastline, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return SexpNull, err
		}
		n := len(lastline)
		if err == io.EOF && n == 0 {
			break
		}
		if n > 0 {
			if lastline[n-1] == '\n' {
				a = append(a, SexpStr(string(lastline[:n-1])))
			} else {
				a = append(a, SexpStr(string(lastline)))
			}
			lineNum += 1
		}

		if err == io.EOF {
			break
		}
	}

	VPrintf("read %d lines\n", lineNum)
	return SexpArray(a), nil
}
