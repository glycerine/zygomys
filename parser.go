package main

import (
	"errors"
	"strconv"
)

const SliceDefaultCap = 10

var SymbolTable map[string]SexpSymbol = make(map[string]SexpSymbol)
var NextSymbolNum int = 0

type Sexp interface {
	SexpString() string
}

type SexpNullValue int
func (null SexpNullValue) SexpString() string {
	return "()"
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

const SexpNull SexpNullValue = 0

type SexpArray []Sexp
type SexpInt int8
type SexpUint uint8

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

func ParseList(tokens []Token) (Sexp, []Token, error) {
	if tokens[0].typ == TokenRParen {
		return SexpNull, tokens[1:], nil
	}

	var start SexpPair
	var expr Sexp
	var err error

	expr, tokens, err = ParseExpression(tokens)
	if err != nil {
		return SexpNull, tokens, err
	}

	start.head = expr

	if tokens[0].typ == TokenDot {
		expr, tokens, err = ParseExpression(tokens[1:])
		if err != nil {
			return SexpNull, tokens, err
		}
		if tokens[0].typ != TokenRParen {
			return SexpNull, tokens,
			       errors.New("extra value in dotted pair")
		}
		start.tail = expr
		return start, tokens[1:], nil
	}

	expr, tokens, err = ParseList(tokens)
	if (err != nil) {
		return start, tokens, err
	}
	start.tail = expr

	return start, tokens, nil
}

func ParseArray(tokens []Token) (Sexp, []Token, error) {
	arr := make([]Sexp, 0, SliceDefaultCap)
	var expr Sexp
	var err error

	for tokens[0].typ != TokenRSquare {
		expr, tokens, err = ParseExpression(tokens)
		if err != nil {
			return SexpNull, tokens, err
		}
		arr = append(arr, expr)
	}

	return SexpArray(arr), tokens[1:], nil
}

func ParseExpression(tokens []Token) (Sexp, []Token, error) {
	var expr Sexp
	var err error
	tok := tokens[0]
	switch (tok.typ) {
	case TokenLParen:
		return ParseList(tokens[1:])
	case TokenLSquare:
		return ParseArray(tokens[1:])
	case TokenQuote:
		expr, tokens, err = ParseExpression(tokens[1:])
		if err != nil {
			return SexpNull, tokens, err
		}
		return MakeList([]Sexp{MakeSymbol("quote"), expr}),
			tokens, nil
	case TokenSymbol:
		return MakeSymbol(tok.str), tokens[1:], nil
	case TokenDecimal:
		i, err := strconv.ParseInt(tok.str, 10, 8)
		if err != nil {
			return SexpNull, tokens, err
		}
		return SexpInt(i), tokens[1:], nil
	case TokenHex:
		i, err := strconv.ParseUint(tok.str, 16, 8)
		if err != nil {
			return SexpNull, tokens, err
		}
		return SexpUint(i), tokens[1:], nil
	case TokenBinary:
		i, err := strconv.ParseUint(tok.str, 2, 8)
		if err != nil {
			return SexpNull, tokens, err
		}
		return SexpUint(i), tokens[1:], nil
	case TokenChar:
		return SexpUint(tok.str[0]), tokens[1:], nil
	}
	return SexpNull, tokens, errors.New("Invalid syntax")
}

func ParseTokens(tokens []Token) ([]Sexp, error) {
	expressions := make([]Sexp, 0, SliceDefaultCap)

	for len(tokens) > 0 {
		var expr Sexp
		var err error
		expr, tokens, err = ParseExpression(tokens)
		if err != nil {
			return expressions, err
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}
