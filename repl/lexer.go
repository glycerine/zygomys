package zygo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"
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
	TokenDollar
	TokenEnd
)

type Token struct {
	typ TokenType
	str string
}

var EndTk = Token{typ: TokenEnd}

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
		return t.str
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
	case TokenDollar:
		return "$"
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
	parser *Parser
	state  LexerState
	tokens []Token
	buffer *bytes.Buffer

	stream   io.RuneScanner
	next     []io.RuneScanner
	linenum  int
	finished bool

	Ready         chan bool
	Done          chan bool
	reqStop       chan bool
	AddInput      chan io.RuneScanner
	ReqReset      chan io.RuneScanner
	NextTokenCh   chan Token
	PeekTokenCh   chan Token
	PingForTokens chan bool

	mut     sync.Mutex
	stopped bool
	peek    Token
}

func NewLexer(p *Parser) *Lexer {
	return &Lexer{
		parser:   p,
		tokens:   make([]Token, 0, 10),
		buffer:   new(bytes.Buffer),
		state:    LexerNormal,
		linenum:  1,
		finished: false,

		Ready:         make(chan bool),
		Done:          make(chan bool),
		reqStop:       make(chan bool),
		ReqReset:      make(chan io.RuneScanner),
		AddInput:      make(chan io.RuneScanner),
		NextTokenCh:   make(chan Token),
		PeekTokenCh:   make(chan Token),
		PingForTokens: make(chan bool, 10),
	}
}

func (p *Lexer) Stop() error {
	p.mut.Lock()
	defer p.mut.Unlock()
	if p.stopped {
		return nil
	}
	p.stopped = true
	close(p.reqStop)
	<-p.Done
	return nil
}

func (p *Lexer) Start() {
	var err error
	var advance bool
	go func() {
		close(p.Ready)
		for {
			advance = true
			if p.stream != nil {
				p.peek, err = p.PeekNextToken()
				if err != nil {
					p.peek = EndTk
					advance = false
				}
			} else {
				W("Lexer has no stream!?! Arg!\n")
			}

			W("Lexer starting select with p.peek = %s\n", p.peek)
			select {
			case <-p.reqStop:
				W("Lexer got reqStop\n")
				close(p.Done)
				return
			case input := <-p.ReqReset:
				W("Lexer got ReqReset\n")
				p.Reset()
				p.AddNextStream(input)
				W("lexer did AddNextStream\n")
			case input := <-p.AddInput:
				W("Lexer got AddInput with %p\n", input)
				p.AddNextStream(input)
			case p.NextTokenChAvail() <- p.peek:
				W("Lexer got NextTokenCh send\n")
				if advance {
					p.tokens = p.tokens[1:]
				}
			case p.PeekTokenChAvail() <- p.peek:
				W("Lexer got PeekTokenCh send\n")
			case <-p.PingForTokens:
				W("got PingForTokens!\n")
			}
		}
	}()
}

func (p *Lexer) NextTokenChAvail() chan Token {
	if len(p.tokens) > 0 {
		return p.NextTokenCh
	}
	return nil
}

func (p *Lexer) PeekTokenChAvail() chan Token {
	return p.PeekTokenCh
}

func (lexer *Lexer) Linenum() int {
	return lexer.linenum
}

func (lex *Lexer) Reset() {
	lex.stream = nil
	lex.tokens = lex.tokens[:0]
	lex.state = LexerNormal
	lex.linenum = 1
	lex.finished = false
	lex.buffer.Reset()
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

	// SymbolRegex = regexp.MustCompile("^[^'#]+$")
	// Symbols cannot contain whitespace nor `~`, `@`, `(`, `)`, `[`, `]`,
	// `{`, `}`, `'`, `#`, `:`, `^`, `\`, `|`, `%`, `"`, `;`
	// Nor, obviously, can they contain backticks, "`".
	// '$' is always a symbol on its own, handled specially.
	SymbolRegex = regexp.MustCompile(`^[^'#:;\\~@\[\]{}\^|"()%]+$`)
	CharRegex   = regexp.MustCompile("^#\\\\?.$")
	FloatRegex  = regexp.MustCompile("^-?([0-9]+\\.[0-9]*)|(\\.[0-9]+)|([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))$")
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
	if atom == "$" {
		return x.Token(TokenSymbol, "$"), nil
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

	return x.EmptyToken(), fmt.Errorf("Unrecognized atom: '%s'", atom)
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
	return EndTk
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

	// $ is always a token and symbol on its own, there
	// is implicit whitespace around it.
	if r == '$' {
		if lexer.buffer.Len() > 0 {
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
		}
		lexer.tokens = append(lexer.tokens, lexer.Token(TokenDollar, "$"))
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

func (lexer *Lexer) PeekNextToken() (tok Token, err error) {
	Q("\n in PeekNextToken()\n")
	defer func() {
		Q("\n done with PeekNextToken() -> returning tok='%v', err=%v. tok='%#v'. tok==EndTk? %v\n",
			tok, err, tok, tok == EndTk)
	}()
	//	if lexer.finished {
	//		return EndTk, nil
	//	}
	if lexer.stream == nil {
		if !lexer.PromoteNextStream() {
			return EndTk, nil
		}
	}

	for len(lexer.tokens) == 0 {
		r, _, err := lexer.stream.ReadRune()
		if err != nil {
			if lexer.PromoteNextStream() {
				continue
			} else {
				// to be continued...
				/*
					lexer.finished = true
						if lexer.buffer.Len() > 0 {
							lexer.dumpBuffer()
							return lexer.tokens[0], nil
						}
				*/
				return EndTk, nil
			}
		}

		err = lexer.LexNextRune(r)
		if err != nil {
			return EndTk, err
		}
	}

	tok = lexer.tokens[0]
	return tok, nil
}

func (lexer *Lexer) GetNextToken() (tok Token, err error) {
	Q("\n in GetNextToken()\n")
	defer func() {
		Q("\n done with GetNextToken() -> returning tok='%v', err=%v.\n", tok, err)
	}()
	tok, err = lexer.PeekNextToken()
	if err != nil || tok.typ == TokenEnd {
		return EndTk, err
	}
	lexer.tokens = lexer.tokens[1:]
	return tok, nil
}

func (lex *Lexer) PromoteNextStream() (ok bool) {
	Q("entering PromoteNextStream()!\n")
	defer func() {
		Q("done with PromoteNextStream, promoted=%v\n", ok)
	}()
	if len(lex.next) == 0 {
		return false
	}
	Q("Promoting next stream!\n")
	lex.stream = lex.next[0]
	lex.next = lex.next[1:]
	return true
}

func (lex *Lexer) AddNextStream(s io.RuneScanner) {
	// in case we still have input available,
	// save new stuff for later
	lex.next = append(lex.next, s)

	lex.finished = false

	if lex.stream == nil {
		lex.PromoteNextStream()
	} else {
		_, _, err := lex.stream.ReadRune()
		if err == nil {
			lex.stream.UnreadRune()
			// still have input available
			return
		} else {
			lex.PromoteNextStream()
		}
	}
}
