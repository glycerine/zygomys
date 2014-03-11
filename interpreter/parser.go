package glisp

import (
	"errors"
	"strconv"
)

type Parser struct {
	lexer *Lexer
	env   *Glisp
}

var UnexpectedEnd error = errors.New("Unexpected end of input")

const SliceDefaultCap = 10

func ParseList(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
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

	expr, err := ParseExpression(parser)
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
		expr, err = ParseExpression(parser)
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

	expr, err = ParseList(parser)
	if err != nil {
		return start, err
	}
	start.tail = expr

	return start, nil
}

func ParseArray(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
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

		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	return SexpArray(arr), nil
}

func ParseExpression(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
	env := parser.env
	tok, err := lexer.GetNextToken()
	if err != nil {
		return SexpEnd, err
	}

	switch tok.typ {
	case TokenLParen:
		return ParseList(parser)
	case TokenLSquare:
		return ParseArray(parser)
	case TokenQuote:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("quote"), expr}), nil
	case TokenSymbol:
		return env.MakeSymbol(tok.str), nil
	case TokenBool:
		return SexpBool(tok.str == "true"), nil
	case TokenDecimal:
		i, err := strconv.ParseInt(tok.str, 10, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpInt(i), nil
	case TokenHex:
		i, err := strconv.ParseInt(tok.str, 16, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpInt(i), nil
	case TokenBinary:
		i, err := strconv.ParseInt(tok.str, 2, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return SexpInt(i), nil
	case TokenChar:
		return SexpChar(tok.str[0]), nil
	case TokenString:
		return SexpStr(tok.str), nil
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

func ParseTokens(env *Glisp, lexer *Lexer) ([]Sexp, error) {
	expressions := make([]Sexp, 0, SliceDefaultCap)
	parser := Parser{lexer, env}

	for {
		expr, err := ParseExpression(&parser)
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
