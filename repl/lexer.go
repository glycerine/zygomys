package zygo

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
	TokenLCurly
	TokenRCurly
	TokenDot
	TokenQuote
	TokenBacktick
	TokenTilde
	TokenTildeAt
	TokenSymbol
	TokenBool
	TokenDecimal
	TokenHex
	TokenOct
	TokenBinary
	TokenFloat
	TokenChar
	TokenString
	TokenCaret
	TokenColonOperator
	TokenThreadingOperator
	TokenBackslash
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
	case TokenLCurly:
		return "{"
	case TokenRCurly:
		return "}"
	case TokenDot:
		return "."
	case TokenQuote:
		return "'"
	case TokenBacktick:
		return "`"
	case TokenCaret:
		return "^"
	case TokenTilde:
		return "~"
	case TokenTildeAt:
		return "~@"
	case TokenHex:
		return "0x" + t.str
	case TokenOct:
		return "0o" + t.str
	case TokenBinary:
		return "0b" + t.str
	case TokenChar:
		quoted := strconv.Quote(t.str)
		return "#" + quoted[1:len(quoted)-1]
	case TokenColonOperator:
		return ":"
	case TokenThreadingOperator:
		return "->"
	case TokenBackslash:
		return "\\"
	}
	return t.str
}

type LexerState int

const (
	LexerNormal LexerState = iota
	LexerComment
	LexerStrLit
	LexerStrEscaped
	LexerUnquote
	LexerBacktickString
)

type Lexer struct {
	state    LexerState
	tokens   []Token
	buffer   *bytes.Buffer
	stream   io.RuneReader
	linenum  int
	finished bool
}

func (lex *Lexer) EmptyToken() Token {
	return Token{}
}

func (lex *Lexer) Token(typ TokenType, str string) Token {
	t := Token{
		typ: typ,
		str: str,
	}
	return t
}

var (
	BoolRegex    = regexp.MustCompile("^(true|false)$")
	DecimalRegex = regexp.MustCompile("^-?[0-9]+$")
	HexRegex     = regexp.MustCompile("^0x[0-9a-fA-F]+$")
	OctRegex     = regexp.MustCompile("^0o[0-7]+$")
	BinaryRegex  = regexp.MustCompile("^0b[01]+$")
	SymbolRegex  = regexp.MustCompile("^[^'#]+$")
	CharRegex    = regexp.MustCompile("^#\\\\?.$")
	FloatRegex   = regexp.MustCompile("^-?([0-9]+\\.[0-9]*)|(\\.[0-9]+)|([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))$")
)

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

func (x *Lexer) DecodeAtom(atom string) (Token, error) {
	if atom == "." {
		return x.Token(TokenDot, ""), nil
	}
	if atom == "\\" {
		return x.Token(TokenBackslash, ""), nil
	}
	if BoolRegex.MatchString(atom) {
		return x.Token(TokenBool, atom), nil
	}
	if DecimalRegex.MatchString(atom) {
		return x.Token(TokenDecimal, atom), nil
	}
	if HexRegex.MatchString(atom) {
		return x.Token(TokenHex, atom[2:]), nil
	}
	if OctRegex.MatchString(atom) {
		return x.Token(TokenOct, atom[2:]), nil
	}
	if BinaryRegex.MatchString(atom) {
		return x.Token(TokenBinary, atom[2:]), nil
	}
	if FloatRegex.MatchString(atom) {
		return x.Token(TokenFloat, atom), nil
	}
	if SymbolRegex.MatchString(atom) {
		return x.Token(TokenSymbol, atom), nil
	}
	if CharRegex.MatchString(atom) {
		char, err := DecodeChar(atom)
		if err != nil {
			return x.EmptyToken(), err
		}
		return x.Token(TokenChar, char), nil
	}

	return x.EmptyToken(), errors.New("Unrecognized atom")
}

func (lexer *Lexer) dumpBuffer() error {
	n := lexer.buffer.Len()
	if n <= 0 {
		return nil
	}

	tok, err := lexer.DecodeAtom(lexer.buffer.String())
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
	lexer.tokens = append(lexer.tokens, lexer.Token(TokenString, str))
}

func (x *Lexer) DecodeBrace(brace rune) Token {
	switch brace {
	case '(':
		return x.Token(TokenLParen, "")
	case ')':
		return x.Token(TokenRParen, "")
	case '[':
		return x.Token(TokenLSquare, "")
	case ']':
		return x.Token(TokenRSquare, "")
	case '{':
		return x.Token(TokenLCurly, "")
	case '}':
		return x.Token(TokenRCurly, "")
	}
	return x.Token(TokenEnd, "")
}

func (lexer *Lexer) LexNextRune(r rune) error {
	if lexer.state == LexerComment {
		if r == '\n' {
			lexer.state = LexerNormal
		}
		return nil
	}
	if lexer.state == LexerBacktickString {
		if r == '`' {
			lexer.dumpString()
			lexer.state = LexerNormal
			return nil
		}
		lexer.buffer.WriteRune(r)
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
	if lexer.state == LexerUnquote {
		if r == '@' {
			lexer.tokens = append(
				lexer.tokens, lexer.Token(TokenTildeAt, ""))
		} else {
			lexer.tokens = append(
				lexer.tokens, lexer.Token(TokenTilde, ""))
			lexer.buffer.WriteRune(r)
		}
		lexer.state = LexerNormal
		return nil
	}
	if r == '`' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected backtick")
		}
		lexer.state = LexerBacktickString
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

	// colon terminates a keyword symbol, e.g. mykey: "myvalue"; mykey is the symbol
	if r == ':' {
		if lexer.buffer.Len() == 0 {
			lexer.tokens = append(lexer.tokens, lexer.Token(TokenColonOperator, ":"))
			return nil
		}
		// but still allow ':' to be a token terminator at the end of a word.
		lexer.tokens = append(lexer.tokens, lexer.Token(TokenQuote, ""))
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		return nil
	}

	if r == '\'' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote")
		}
		lexer.tokens = append(lexer.tokens, lexer.Token(TokenQuote, ""))
		return nil
	}

	// caret '^' replaces backtick '`' as the start of a macro template, so
	// we can use `` as in Go for verbatim strings (strings with newlines, etc).
	if r == '^' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected ^ caret")
		}
		lexer.tokens = append(lexer.tokens, lexer.Token(TokenCaret, ""))
		return nil
	}

	if r == '~' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected tilde")
		}
		lexer.state = LexerUnquote
		return nil
	}

	if r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' {
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		lexer.tokens = append(lexer.tokens, lexer.DecodeBrace(r))
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
		return lexer.Token(TokenEnd, ""), nil
	}

	for len(lexer.tokens) == 0 {
		r, _, err := lexer.stream.ReadRune()

		if err != nil {
			lexer.finished = true
			if lexer.buffer.Len() > 0 {
				lexer.dumpBuffer()
				return lexer.tokens[0], nil
			}
			return lexer.Token(TokenEnd, ""), nil
		}

		err = lexer.LexNextRune(r)
		if err != nil {
			return lexer.Token(TokenEnd, ""), err
		}
	}

	tok := lexer.tokens[0]
	return tok, nil
}

func (lexer *Lexer) GetNextToken() (Token, error) {
	tok, err := lexer.PeekNextToken()
	if err != nil || tok.typ == TokenEnd {
		return lexer.Token(TokenEnd, ""), err
	}
	lexer.tokens = lexer.tokens[1:]
	return tok, nil
}

func NewLexerFromStream(stream io.RuneReader) *Lexer {
	return &Lexer{
		tokens:   make([]Token, 0, 10),
		buffer:   new(bytes.Buffer),
		state:    LexerNormal,
		stream:   stream,
		linenum:  1,
		finished: false,
	}
}

func (lexer *Lexer) Linenum() int {
	return lexer.linenum
}
