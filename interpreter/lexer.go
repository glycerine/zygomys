package glisp

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"unicode/utf8"
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
	TokenBool
	TokenDecimal
	TokenHex
	TokenBinary
	TokenFloat
	TokenChar
	TokenString
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

type LexerState int

const (
	LexerNormal LexerState = iota
	LexerComment
	LexerStrLit
	LexerStrEscaped
)

type Lexer struct {
	state    LexerState
	tokens   []Token
	buffer   *bytes.Buffer
	stream   io.RuneReader
	linenum  int
	finished bool
}

var BoolRegex *regexp.Regexp
var DecimalRegex *regexp.Regexp
var HexRegex *regexp.Regexp
var BinaryRegex *regexp.Regexp
var SymbolRegex *regexp.Regexp
var CharRegex *regexp.Regexp
var FloatRegex *regexp.Regexp
var lexInit = false

func InitLexer() {
	var err error

	BoolRegex, err = regexp.Compile("^(true|false)$")
	if err != nil {
		panic(err)
	}
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
	FloatRegex, err = regexp.Compile("^-?([0-9]+\\.[0-9]*)|(\\.[0-9]+)|([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))$")
	if err != nil {
		panic(err)
	}
	lexInit = true
}

func StringToRunes(str string) []rune {
	b := []byte(str)
	runes := make([]rune, 0)

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)
		b = b[size:]
	}
	return runes
}

func EscapeChar(char rune) (rune, error) {
	switch char {
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 'a':
		return '\a', nil
	case 't':
		return '\t', nil
	case '\\':
		return '\\', nil
	case '"':
		return '"', nil
	case '\'':
		return '\'', nil
	case '#':
		return '#', nil
	}
	return ' ', errors.New("invalid escape sequence")
}

func DecodeChar(atom string) (string, error) {
	runes := StringToRunes(atom)
	if len(runes) == 3 {
		char, err := EscapeChar(runes[2])
		return string(char), err
	}

	if len(runes) == 2 {
		return string(runes[1:2]), nil
	}
	return "", errors.New("not a char literal")
}

func DecodeAtom(atom string) (Token, error) {
	if atom == "." {
		return Token{TokenDot, ""}, nil
	}
	if BoolRegex.MatchString(atom) {
		return Token{TokenBool, atom}, nil
	}
	if DecimalRegex.MatchString(atom) {
		return Token{TokenDecimal, atom}, nil
	}
	if HexRegex.MatchString(atom) {
		return Token{TokenHex, atom[2:]}, nil
	}
	if BinaryRegex.MatchString(atom) {
		return Token{TokenBinary, atom[2:]}, nil
	}
	if FloatRegex.MatchString(atom) {
		return Token{TokenFloat, atom}, nil
	}
	if SymbolRegex.MatchString(atom) {
		return Token{TokenSymbol, atom}, nil
	}
	if CharRegex.MatchString(atom) {
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

func (lexer *Lexer) dumpString() {
	str := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.tokens = append(lexer.tokens, Token{TokenString, str})
}

func DecodeBrace(brace rune) Token {
	switch brace {
	case '(':
		return Token{TokenLParen, ""}
	case ')':
		return Token{TokenRParen, ""}
	case '[':
		return Token{TokenLSquare, ""}
	case ']':
		return Token{TokenRSquare, ""}
	}
	return Token{TokenEnd, ""}
}

func (lexer *Lexer) LexNextRune(r rune) error {
	if lexer.state == LexerComment {
		if r == '\n' {
			lexer.state = LexerNormal
		}
		return nil
	}
	if lexer.state == LexerStrLit {
		if r == '\\' {
			lexer.state = LexerStrEscaped
			return nil
		}
		if r == '"' {
			lexer.dumpString()
			lexer.state = LexerNormal
			return nil
		}
		lexer.buffer.WriteRune(r)
		return nil
	}
	if lexer.state == LexerStrEscaped {
		char, err := EscapeChar(r)
		if err != nil {
			return err
		}
		lexer.buffer.WriteRune(char)
		lexer.state = LexerStrLit
		return nil
	}

	if r == '"' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote")
		}
		lexer.state = LexerStrLit
		return nil
	}

	if r == ';' {
		lexer.state = LexerComment
		return nil
	}

	if r == '\'' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote")
		}
		lexer.tokens = append(lexer.tokens, Token{TokenQuote, ""})
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
		if r == '\n' {
			lexer.linenum++
		}
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		return nil
	}

	_, err := lexer.buffer.WriteRune(r)
	if err != nil {
		return err
	}
	return nil
}

func (lexer *Lexer) PeekNextToken() (Token, error) {
	if lexer.finished {
		return Token{TokenEnd, ""}, nil
	}
	for len(lexer.tokens) == 0 {
		r, _, err := lexer.stream.ReadRune()
		if err != nil {
			lexer.finished = true
			if lexer.buffer.Len() > 0 {
				lexer.dumpBuffer()
				return lexer.tokens[0], nil
			}
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
	lexer.state = LexerNormal
	lexer.stream = stream
	lexer.linenum = 1
	lexer.finished = false

	return lexer
}

func (lexer *Lexer) Linenum() int {
	return lexer.linenum
}
