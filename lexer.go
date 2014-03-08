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
	TokenEnd
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

type Lexer struct {
	comment bool
	tokens []Token
	buffer *bytes.Buffer
	stream io.RuneReader
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

func (lexer *Lexer) dumpBuffer() error {
	if lexer.buffer.Len() <= 0 {
		return nil
	}

	tok, err := DecodeAtom(lexer.buffer.String())
	if err != nil {
		return err
	}

	lexer.buffer.Reset()
	lexer.tokens = append(lexer.tokens, tok)
	return nil
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

func (lexer *Lexer) LexNextRune(r rune) error {
	if lexer.comment {
		if r == '\n' {
			lexer.comment = false
		}
		return nil
	}

	if r == '.' || r == '\'' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote or dot")
		}
		lexer.tokens = append(lexer.tokens, DecodePunctuation(r))
		return nil
	}

	if r == '(' || r == ')' || r == '[' || r == ']' {
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		lexer.tokens = append(lexer.tokens, DecodeBrace(r))
		return nil
	}
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		return nil
	}
	if r == ';' {
		lexer.comment = true
		return nil
	}

	_, err := lexer.buffer.WriteRune(r)
	if err != nil {
		return err
	}
	return nil
}

func (lexer *Lexer) PeekNextToken() (Token, error) {
	for len(lexer.tokens) == 0 {
		r, _, err := lexer.stream.ReadRune()
		if err != nil {
			return Token{TokenEnd, ""}, nil
		}

		err = lexer.LexNextRune(r)
		if err != nil {
			return Token{TokenEnd, ""}, err
		}
	}

	tok := lexer.tokens[0]
	return tok, nil
}

func (lexer *Lexer) GetNextToken() (Token, error) {
	tok, err := lexer.PeekNextToken()
	if err != nil || tok.typ == TokenEnd {
		return Token{TokenEnd, ""}, err
	}
	lexer.tokens = lexer.tokens[1:]
	return tok, nil
}

func NewLexerFromStream(stream io.RuneReader) *Lexer {
	if !lexInit {
		InitLexer()
	}

	lexer := new(Lexer)

	lexer.tokens = make([]Token, 0, 10)
	lexer.buffer = new(bytes.Buffer)
	lexer.comment = false
	lexer.stream = stream

	return lexer
}
