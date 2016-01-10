package zygo

import (
	"fmt"
	"os/exec"
	"strings"
)

// system: ($) is macro. shell out, return the combined output.
func SystemFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) == 0 {
		return SexpNull, WrongNargs
	}
	cmd := ""
	stringArgs := []string{}
	switch c := args[0].(type) {
	case SexpStr:
		many := strings.Split(string(c), " ")
		cmd = string(many[0])
		stringArgs = append(stringArgs, many[1:]...)
	default:
		return SexpNull, fmt.Errorf("arguments to system must be strings")
	}

	for _, word := range args[1:] {
		switch s := word.(type) {
		case SexpStr:
			many := strings.Split(string(s), " ")
			stringArgs = append(stringArgs, many...)
		default:
			return SexpNull, fmt.Errorf("arguments to system must be strings")
		}
	}

	out, err := exec.Command(cmd, stringArgs...).CombinedOutput()
	if err != nil {
		return SexpNull, err
	}
	return SexpStr(string(out)), nil
}
