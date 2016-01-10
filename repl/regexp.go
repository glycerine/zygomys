package zygo

import (
	"errors"
	"fmt"
	"regexp"
)

type SexpRegexp regexp.Regexp

func (re SexpRegexp) SexpString() string {
	r := regexp.Regexp(re)
	return fmt.Sprintf(`(regexp-compile "%v")`, r.String())
}

func regexpFindIndex(
	needle regexp.Regexp, haystack string) (Sexp, error) {

	loc := needle.FindStringIndex(haystack)

	arr := make([]Sexp, len(loc))
	for i := range arr {
		arr[i] = Sexp(SexpInt(loc[i]))
	}

	return SexpArray(arr), nil
}

func RegexpFind(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var haystack string
	switch t := args[1].(type) {
	case SexpStr:
		haystack = string(t)
	default:
		return SexpNull,
			errors.New(fmt.Sprintf("2nd argument of %v should be a string", name))
	}

	var needle regexp.Regexp
	switch t := args[0].(type) {
	case SexpRegexp:
		needle = regexp.Regexp(t)
	default:
		return SexpNull,
			errors.New(fmt.Sprintf("1st argument of %v should be a compiled regular expression", name))
	}

	switch name {
	case "regexp-find":
		str := needle.FindString(haystack)
		return SexpStr(str), nil
	case "regexp-find-index":
		return regexpFindIndex(needle, haystack)
	case "regexp-match":
		matches := needle.MatchString(haystack)
		return SexpBool(matches), nil
	}

	return SexpNull, errors.New("unknown function")
}

func RegexpCompile(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var re string
	switch t := args[0].(type) {
	case SexpStr:
		re = string(t)
	default:
		return SexpNull,
			errors.New("argument of regexp-compile should be a string")
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return SexpNull, errors.New(
			fmt.Sprintf("error during regexp-compile: '%v'", err))
	}

	return Sexp(SexpRegexp(*r)), nil
}

func (env *Glisp) ImportRegex() {
	env.AddFunction("regexp-compile", RegexpCompile)
	env.AddFunction("regexp-find-index", RegexpFind)
	env.AddFunction("regexp-find", RegexpFind)
	env.AddFunction("regexp-match", RegexpFind)
}
