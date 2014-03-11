package glisp

import (
	"reflect"
	"strconv"
	"strings"
)

type Sexp interface {
	SexpString() string
}

type SexpSentinel int

const (
	SexpNull SexpSentinel = iota
	SexpEnd
)

func (sent SexpSentinel) SexpString() string {
	if sent == SexpNull {
		return "()"
	}
	if sent == SexpEnd {
		return "End"
	}

	return ""
}

type SexpPair struct {
	head Sexp
	tail Sexp
}

func (pair SexpPair) SexpString() string {
	str := "("

	for {
		switch pair.tail.(type) {
		case SexpPair:
			str += pair.head.SexpString() + " "
			pair = pair.tail.(SexpPair)
			continue
		}
		break
	}

	str += pair.head.SexpString()

	if pair.tail == SexpNull {
		str += ")"
	} else {
		str += " . " + pair.tail.SexpString() + ")"
	}

	return str
}

type SexpArray []Sexp
type SexpInt int
type SexpBool bool
type SexpFloat float64
type SexpChar rune
type SexpStr string

var SexpIntSize = reflect.TypeOf(SexpInt(0)).Bits()
var SexpFloatSize = reflect.TypeOf(SexpFloat(0.0)).Bits()

func (arr SexpArray) SexpString() string {
	if len(arr) == 0 {
		return "[]"
	}

	str := "[" + arr[0].SexpString()
	for _, sexp := range arr[1:] {
		str += " " + sexp.SexpString()
	}
	str += "]"
	return str
}

func (b SexpBool) SexpString() string {
	if b {
		return "true"
	}
	return "false"
}

func (i SexpInt) SexpString() string {
	return strconv.Itoa(int(i))
}

func (f SexpFloat) SexpString() string {
	return strconv.FormatFloat(float64(f), 'g', 5, SexpFloatSize)
}

func (c SexpChar) SexpString() string {
	return "#" + strings.Trim(strconv.QuoteRune(rune(c)), "'")
}

func (s SexpStr) SexpString() string {
	return strconv.Quote(string(s))
}

type SexpSymbol struct {
	name   string
	number int
}

func (sym SexpSymbol) SexpString() string {
	return sym.name
}

type SexpFunction struct {
	name    string
	user    bool
	nargs   int
	fun     GlispFunction
	userfun GlispUserFunction
}

func (sf SexpFunction) SexpString() string {
	return "fn [" + sf.name + "]"
}

func IsTruthy(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpBool:
		return bool(e)
	case SexpInt:
		return e != 0
	case SexpChar:
		return e != 0
	case SexpSentinel:
		return e != SexpNull
	}
	return true
}
