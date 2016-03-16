package zygo

import (
	"errors"
	"fmt"
	"regexp"
)

type SexpRegexp regexp.Regexp

func (re *SexpRegexp) SexpString(indent int) string {
	r := (*regexp.Regexp)(re)
	return fmt.Sprintf(`(regexpCompile "%v")`, r.String())
}

func (r *SexpRegexp) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func regexpFindIndex(
	needle *regexp.Regexp, haystack string) (Sexp, error) {

	loc := needle.FindStringIndex(haystack)

	arr := make([]Sexp, len(loc))
	for i := range arr {
		arr[i] = Sexp(&SexpInt{Val: int64(loc[i])})
	}

	return &SexpArray{Val: arr}, nil
}

func RegexpFind(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var haystack string
	switch t := args[1].(type) {
	case *SexpStr:
		haystack = t.S
	default:
		return SexpNull,
			errors.New(fmt.Sprintf("2nd argument of %v should be a string", name))
	}

	var needle *regexp.Regexp
	switch t := args[0].(type) {
	case *SexpRegexp:
		needle = (*regexp.Regexp)(t)
	default:
		return SexpNull,
			errors.New(fmt.Sprintf("1st argument of %v should be a compiled regular expression", name))
	}

	switch name {
	case "regexpFind":
		str := needle.FindString(haystack)
		return &SexpStr{S: str}, nil
	case "regexpFindIndex":
		return regexpFindIndex(needle, haystack)
	case "regexpMatch":
		matches := needle.MatchString(haystack)
		return &SexpBool{Val: matches}, nil
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
	case *SexpStr:
		re = t.S
	default:
		return SexpNull,
			errors.New("argument of regexpCompile should be a string")
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return SexpNull, errors.New(
			fmt.Sprintf("error during regexpCompile: '%v'", err))
	}

	return Sexp((*SexpRegexp)(r)), nil
}

func (env *Glisp) ImportRegex() {
	env.AddFunction("regexpCompile", RegexpCompile)
	env.AddFunction("regexpFindIndex", RegexpFind)
	env.AddFunction("regexpFind", RegexpFind)
	env.AddFunction("regexpMatch", RegexpFind)
}
