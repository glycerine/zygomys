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
	SexpMarker
)

func (sent SexpSentinel) SexpString() string {
	if sent == SexpNull {
		return "()"
	}
	if sent == SexpEnd {
		return "End"
	}
	if sent == SexpMarker {
		return "Marker"
	}

	return ""
}

type SexpPair struct {
	head Sexp
	tail Sexp
}

func Cons(a Sexp, b Sexp) SexpPair {
	return SexpPair{a, b}
}

func (pair SexpPair) Head() Sexp {
	return pair.head
}

func (pair SexpPair) Tail() Sexp {
	return pair.tail
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
type SexpHash map[int][]SexpPair
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

func (hash SexpHash) SexpString() string {
	str := "{"
	for _, arr := range hash {
		for _, pair := range arr {
			str += pair.head.SexpString() + " "
			str += pair.tail.SexpString() + " "
		}
	}
	if len(str) > 1 {
		return str[:len(str)-1] + "}"
	}
	return str + "}"
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

func (sym SexpSymbol) Name() string {
	return sym.name
}

func (sym SexpSymbol) Number() int {
	return sym.number
}

type SexpFunction struct {
	name       string
	user       bool
	nargs      int
	varargs    bool
	fun        GlispFunction
	userfun    GlispUserFunction
	closeScope *Stack
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
