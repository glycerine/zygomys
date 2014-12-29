package glispext

import (
	"errors"
	"fmt"
	"regexp"

	glisp "github.com/zhemao/glisp/interpreter"
)

type SexpRegexp regexp.Regexp

func (re SexpRegexp) SexpString() string {
	r := regexp.Regexp(re)
	return fmt.Sprintf(`(regexp.Compile "%v")`, r.String())
}

func RegexpFindStringIndex(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) != 2 {
		return glisp.SexpNull, glisp.WrongNargs
	}
	var haystack string
	switch t := args[1].(type) {
	case glisp.SexpStr:
		haystack = string(t)
	default:
		return glisp.SexpNull,
			errors.New("2nd argument of regexp.FindStringIndex should be a string to check against the regexp of the first argument.")
	}

	var needle regexp.Regexp
	switch t := args[0].(type) {
	case SexpRegexp:
		needle = regexp.Regexp(t)
	default:
		return glisp.SexpNull,
			errors.New("1st argument of regexp.FindStringIndex should be a compiled regular expression")
	}

	loc := needle.FindStringIndex(haystack)

	arr := make([]glisp.Sexp, len(loc))
	for i := range arr {
		arr[i] = glisp.Sexp(glisp.SexpInt(loc[i]))
	}

	return glisp.SexpArray(arr), nil
}

func RegexpCompile(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) < 1 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	var re string
	switch t := args[0].(type) {
	case glisp.SexpStr:
		re = string(t)
	default:
		return glisp.SexpNull,
			errors.New("argument of regexp.Compile should be a string")
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return glisp.SexpNull, errors.New(
			fmt.Sprintf("error during regexp.Compile: '%v'", err))
	}

	return glisp.Sexp(SexpRegexp(*r)), nil
}

func ImportRegex(env *glisp.Glisp) {
	env.AddFunction("regexp.Compile", RegexpCompile)
	env.AddFunction("regexp.FindStringIndex", RegexpFindStringIndex)
}
