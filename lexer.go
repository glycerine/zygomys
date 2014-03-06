package main

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
)

type TokenType int

const (
	TokenLParen TokenType = iota
	TokenRParen
	TokenLSquare
	TokenRSquare
	TokenDot
	TokenQuote
	TokenSymbol
	TokenDecimal
	TokenHex
	TokenBinary
	TokenChar
)

type Token struct {
	typ TokenType
	str string
}

func (t Token) String() string {
	switch t.typ {
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenLSquare:
		return "["
	case TokenRSquare:
		return "]"
	case TokenQuote:
		return "'"
	case TokenSymbol:
		return t.str
	case TokenDecimal:
		return t.str
	case TokenHex:
		return "0x" + t.str
	case TokenBinary:
		return "0b" + t.str
	case TokenChar:
		quoted := strconv.Quote(t.str)
		return "#" + quoted[1:len(quoted)-1]
	}
	return ""
}

var DecimalRegex *regexp.Regexp
var HexRegex *regexp.Regexp
var BinaryRegex *regexp.Regexp
var SymbolRegex *regexp.Regexp
var CharRegex *regexp.Regexp
var lexInit = false

func InitLexer() {
	var err error

	DecimalRegex, err = regexp.Compile("^-?[0-9]+$")
	if err != nil {
		panic(err)
	}
	HexRegex, err = regexp.Compile("^0x[0-9a-fA-F]+$")
	if err != nil {
		panic(err)
	}
	BinaryRegex, err = regexp.Compile("^0b[01]+$")
	if err != nil {
		panic(err)
	}
	SymbolRegex, err = regexp.Compile("^[^'#]+$")
	if err != nil {
		panic(err)
	}
	CharRegex, err = regexp.Compile("^#\\\\?.$")
	if err != nil {
		panic(err)
	}
	lexInit = true
}

func DecodeChar(atom string) (string, error) {
	if len(atom) == 3 {
		switch atom[2] {
		case 'n':
			return "\n", nil
		case 'r':
			return "\r", nil
		case 'a':
			return "\a", nil
		case '#':
			return "#", nil
		default:
			return "", errors.New("invalid escape sequence")
		}
	}

	return atom[1:2], nil
}

func DecodeAtom(atom string) (Token, error) {
	if DecimalRegex.MatchString(atom) {
		return Token{TokenDecimal, atom}, nil
	} else if HexRegex.MatchString(atom) {
		return Token{TokenHex, atom[2:]}, nil
	} else if BinaryRegex.MatchString(atom) {
		return Token{TokenBinary, atom[2:]}, nil
	} else if SymbolRegex.MatchString(atom) {
		return Token{TokenSymbol, atom}, nil
	} else if CharRegex.MatchString(atom) {
		char, err := DecodeChar(atom)
		if err != nil {
			return Token{}, err
		}
		return Token{TokenChar, char}, nil
	}

	return Token{}, errors.New("Unrecognized atom")
}

func dumpBuffer(tokens []Token, buffer *bytes.Buffer) ([]Token, error) {
	if buffer.Len() <= 0 {
		return tokens, nil
	}

	tok, err := DecodeAtom(buffer.String())
	if err != nil {
		return tokens, err
	}
	buffer.Reset()
	return append(tokens, tok), nil
}

func DecodeBrace(brace rune) Token {
	switch brace {
	case '(':
		return Token{TokenLParen, ""}
	case ')':
		return Token{TokenRParen, ""}
	case '[':
		return Token{TokenLSquare, ""}
	default:
		return Token{TokenRSquare, ""}
	}
}

func DecodePunctuation(punct rune) Token {
	switch punct {
	case '\'':
		return Token{TokenQuote, ""}
	case '.':
		return Token{TokenDot, ""}
	}
	panic("Punctuation is not a quote or dot")
}

func LexNextRune(buffer *bytes.Buffer, tokens []Token, r rune, comment bool) ([]Token, bool, error) {
	var err error

	if comment {
		if r == '\n' {
			return tokens, false, nil
		}
		return tokens, true, nil
	}

	if r == '.' || r == '\'' {
		if buffer.Len() > 0 {
			return tokens, false,
				errors.New("Unexpected quote or dot")
		}
		tokens = append(tokens, DecodePunctuation(r))
		return tokens, false, nil
	}

	if r == '(' || r == ')' || r == '[' || r == ']' {
		tokens, err = dumpBuffer(tokens, buffer)
		if err != nil {
			return tokens, false, err
		}
		tokens = append(tokens, DecodeBrace(r))
		return tokens, false, nil
	}
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		tokens, err = dumpBuffer(tokens, buffer)
		if err != nil {
			return tokens, false, err
		}
		return tokens, false, nil
	}
	if r == ';' {
		comment = true
		return tokens, true, nil
	}

	_, err = buffer.WriteRune(r)
	if err != nil {
		return tokens, false, err
	}
	return tokens, false, nil
}

func LexStream(stream io.RuneReader) ([]Token, error) {
	if !lexInit {
		InitLexer()
	}

	tokens := make([]Token, 0, 20)
	comment := false
	buffer := new(bytes.Buffer)

	for {
		r, _, err := stream.ReadRune()
		if err != nil {
			break
		}

		tokens, comment, err = LexNextRune(buffer, tokens, r, comment)
		if err != nil {
			return tokens, err
		}
	}

	return tokens, nil
}
