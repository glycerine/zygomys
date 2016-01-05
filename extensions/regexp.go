package gdslext

import (
	"errors"
	"fmt"
	"regexp"

	gdsl "github.com/glycerine/godiesel/interpreter"
)

type SexpRegexp regexp.Regexp

func (re SexpRegexp) SexpString() string {
	r := regexp.Regexp(re)
	return fmt.Sprintf(`(regexp-compile "%v")`, r.String())
}

func regexpFindIndex(
	needle regexp.Regexp, haystack string) (gdsl.Sexp, error) {

	loc := needle.FindStringIndex(haystack)

	arr := make([]gdsl.Sexp, len(loc))
	for i := range arr {
		arr[i] = gdsl.Sexp(gdsl.SexpInt(loc[i]))
	}

	return gdsl.SexpArray(arr), nil
}

func RegexpFind(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	if len(args) != 2 {
		return gdsl.SexpNull, gdsl.WrongNargs
	}
	var haystack string
	switch t := args[1].(type) {
	case gdsl.SexpStr:
		haystack = string(t)
	default:
		return gdsl.SexpNull,
			errors.New(fmt.Sprintf("2nd argument of %v should be a string", name))
	}

	var needle regexp.Regexp
	switch t := args[0].(type) {
	case SexpRegexp:
		needle = regexp.Regexp(t)
	default:
		return gdsl.SexpNull,
			errors.New(fmt.Sprintf("1st argument of %v should be a compiled regular expression", name))
	}

	switch name {
	case "regexp-find":
		str := needle.FindString(haystack)
		return gdsl.SexpStr(str), nil
	case "regexp-find-index":
		return regexpFindIndex(needle, haystack)
	case "regexp-match":
		matches := needle.MatchString(haystack)
		return gdsl.SexpBool(matches), nil
	}

	return gdsl.SexpNull, errors.New("unknown function")
}

func RegexpCompile(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	if len(args) < 1 {
		return gdsl.SexpNull, gdsl.WrongNargs
	}

	var re string
	switch t := args[0].(type) {
	case gdsl.SexpStr:
		re = string(t)
	default:
		return gdsl.SexpNull,
			errors.New("argument of regexp-compile should be a string")
	}

	r, err := regexp.Compile(re)

	if err != nil {
		return gdsl.SexpNull, errors.New(
			fmt.Sprintf("error during regexp-compile: '%v'", err))
	}

	return gdsl.Sexp(SexpRegexp(*r)), nil
}

func ImportRegex(env *gdsl.Glisp) {
	env.AddFunction("regexp-compile", RegexpCompile)
	env.AddFunction("regexp-find-index", RegexpFind)
	env.AddFunction("regexp-find", RegexpFind)
	env.AddFunction("regexp-match", RegexpFind)
}
