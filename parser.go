package main

import (
	"errors"
	"strconv"
	"reflect"
	"strings"
)

var UnexpectedEnd error = errors.New("Unexpected end of input")

const SliceDefaultCap = 10

var SymbolTable map[string]SexpSymbol = make(map[string]SexpSymbol)
var NextSymbolNum int = 0

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
type SexpUint uint
type SexpFloat float64
type SexpChar rune

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

func (i SexpInt) SexpString() string {
	return strconv.Itoa(int(i))
}

func (i SexpUint) SexpString() string {
	return strconv.Itoa(int(i))
}

func (f SexpFloat) SexpString() string {
	return strconv.FormatFloat(float64(f), 'g', 5, SexpFloatSize)
}

func (c SexpChar) SexpString() string {
	return "#" + strings.Trim(strconv.QuoteRune(rune(c)), "'")
}

type SexpSymbol struct {
	name   string
	number int
}

func (sym SexpSymbol) SexpString() string {
	return sym.name
}

func MakeSymbol(name string) SexpSymbol {
	symbol := SexpSymbol{name, NextSymbolNum}
	SymbolTable[name] = symbol
	NextSymbolNum += 1
	return symbol
}

func MakeList(expressions []Sexp) Sexp {
	if len(expressions) == 0 {
		return SexpNull
	}

	return SexpPair{expressions[0], MakeList(expressions[1:])}
}

func ParseList(lexer *Lexer) (Sexp, error) {
	tok, err := lexer.PeekNextToken()
	if err != nil {
		return SexpNull, err
	}
	if tok.typ == TokenEnd {
		_, _ = lexer.GetNextToken()
		return SexpEnd, UnexpectedEnd
	}

	if tok.typ == TokenRParen {
		_, _ = lexer.GetNextToken()
		return SexpNull, nil
	}

	var start SexpPair

	expr, err := ParseExpression(lexer)
	if err != nil {
		return SexpNull, err
	}

	start.head = expr

	tok, err = lexer.PeekNextToken()
	if err != nil {
		return SexpNull, err
	}

	if tok.typ == TokenDot {
		// eat up the dot
		_, _ = lexer.GetNextToken()
		expr, err = ParseExpression(lexer)
		if err != nil {
			return SexpNull, err
		}

		// eat up the end paren
		tok, err = lexer.GetNextToken()
		if err != nil {
			return SexpNull, err
		}
		// make sure it was actually an end paren
		if tok.typ != TokenRParen {
			return SexpNull, errors.New("extra value in dotted pair")
		}
		start.tail = expr
		return start, nil
	}

	expr, err = ParseList(lexer)
	if err != nil {
		return start, err
	}
	start.tail = expr

	return start, nil
}

func ParseArray(lexer *Lexer) (Sexp, error) {
	arr := make([]Sexp, 0, SliceDefaultCap)

	for {
		tok, err := lexer.PeekNextToken()
		if err != nil {
			return SexpEnd, err
		}

		if tok.typ == TokenEnd {
			return SexpEnd, UnexpectedEnd
		}

		if tok.typ == TokenRSquare {
			// pop off the ]
			_, _ = lexer.GetNextToken()
			break
		}

		expr, err := ParseExpression(lexer)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	return SexpArray(arr), nil
}

func ParseExpression(lexer *Lexer) (Sexp, error) {
	tok, err := lexer.GetNextToken()
	if err != nil {
		return SexpEnd, err
	}

	switch (tok.typ) {
	case TokenLParen:
		return ParseList(lexer)
	case TokenLSquare:
		return ParseArray(lexer)
	case TokenQuote:
		expr, err := ParseExpression(lexer)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{MakeSymbol("quote"), expr}), nil
	case TokenSymbol:
		return MakeSymbol(tok.str), nil
	case TokenDecimal:
		i, err := strconv.ParseInt(tok.str, 10, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpInt(i), nil
	case TokenHex:
		i, err := strconv.ParseUint(tok.str, 16, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpUint(i), nil
	case TokenBinary:
		i, err := strconv.ParseUint(tok.str, 2, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpUint(i), nil
	case TokenChar:
		return SexpChar(tok.str[0]), nil
	case TokenFloat:
		f, err := strconv.ParseFloat(tok.str, SexpFloatSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpFloat(f), nil
	case TokenEnd:
		return SexpEnd, nil
	}
	return SexpNull, errors.New("Invalid syntax")
}

func ParseTokens(lexer *Lexer) ([]Sexp, error) {
	expressions := make([]Sexp, 0, SliceDefaultCap)

	for {
		expr, err := ParseExpression(lexer)
		if err != nil {
			return expressions, err
		}
		if expr == SexpEnd {
			break
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}
