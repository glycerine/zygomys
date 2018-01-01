package zygo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type TokenType int

const (
	TokenTypeEmpty TokenType = iota
	TokenLParen
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
	TokenDotSymbol
	TokenFreshAssign
	TokenBacktickString
	TokenComment
	TokenBeginBlockComment
	TokenEndBlockComment
	TokenSemicolon
	TokenSymbolColon
	TokenComma
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
		return strconv.Quote(t.str)
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
	LexerNormal         LexerState = iota
	LexerCommentLine               //
	LexerStrLit                    //
	LexerStrEscaped                //
	LexerUnquote                   //
	LexerBacktickString            //
	LexerFreshAssignOrColon
	LexerFirstFwdSlash // could be start of // comment or /*
	LexerCommentBlock
	LexerCommentBlockAsterisk // could be end of block comment */
	LexerBuiltinOperator
)

type Lexer struct {
	parser   *Parser
	state    LexerState
	prevrune rune
	tokens   []Token
	buffer   *bytes.Buffer

	prevToken     Token
	prevPrevToken Token
	stream        io.RuneScanner
	next          []io.RuneScanner
	linenum       int
}

func (lexer *Lexer) AppendToken(tok Token) {
	lexer.tokens = append(lexer.tokens, tok)
	lexer.prevPrevToken = lexer.prevToken
	lexer.prevToken = tok
}

func (lexer *Lexer) PrependToken(tok Token) {
	lexer.tokens = append([]Token{tok}, lexer.tokens...)
}

func NewLexer(p *Parser) *Lexer {
	return &Lexer{
		parser:  p,
		tokens:  make([]Token, 0, 10),
		buffer:  new(bytes.Buffer),
		state:   LexerNormal,
		linenum: 1,
	}
}

func (lexer *Lexer) Linenum() int {
	return lexer.linenum
}

func (lex *Lexer) Reset() {
	lex.stream = nil
	lex.tokens = lex.tokens[:0]
	lex.state = LexerNormal
	lex.linenum = 1
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
	// (Sigil) symbols can begin with #, $, ?, but
	// sigils cannot appear later in any symbol.
	// Symbols cannot contain whitespace nor `~`, `@`, `(`, `)`, `[`, `]`,
	// `{`, `}`, `'`, `#`, `^`, `\`, `|`, `%`, `"`, `;`. They can optionally
	// end in `:`.
	// Nor, obviously, can symbols contain backticks, "`".
	// Symbols cannot start with a number. DotSymbols cannot have a number
	// as the first character after '.'
	SymbolRegex = regexp.MustCompile(`^[#$?]?[^#$?':;\\~@\[\]{}\^|"()%0-9,&][^'#:;\\~@\[\]{}\^|"()%,&*\-]*[:]?$`)
	// dot symbol examples: `.`, `.a`, `.a.b`, `.a.b.c`
	// dot symbol non-examples: `.a.`, `..`
	DotSymbolRegex = regexp.MustCompile(`^[.]$|^([.][^'#:;\\~@\[\]{}\^|"()%.0-9,][^'#:;\\~@\[\]{}\^|"()%.,*+\-]*)+$|^[^'#:;\\~@\[\]{}\^|"()%.0-9,][^'#:;\\~@\[\]{}\^|"()%.,*+\-]*([.][^'#:;\\~@\[\]{}\^|"()%.0-9,][^'#:;\\~@\[\]{}\^|"()%.,*+\-]*)+$`)
	DotPartsRegex  = regexp.MustCompile(`[.]?[^'#:;\\~@\[\]{}\^|"()%.0-9,][^'#:;\\~@\[\]{}\^|"()%.,]*`)
	CharRegex      = regexp.MustCompile("^'\\\\?.'$")
	FloatRegex     = regexp.MustCompile("^-?([0-9]+\\.[0-9]*)$|-?(\\.[0-9]+)$|-?([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))$")
	ComplexRegex   = regexp.MustCompile("^-?([0-9]+\\.[0-9]*)i?$|-?(\\.[0-9]+)i?$|-?([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))i?$")
	BuiltinOpRegex = regexp.MustCompile(`^(\+\+|\-\-|\+=|\-=|=|==|:=|\+|\-|\*|<|>|<=|>=|<-|->|\*=|/=|\*\*|!|!=|<!)$`)
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
	n := len(runes)
	runes = runes[:n-1]
	runes = runes[1:]
	if len(runes) == 2 {
		char, err := EscapeChar(runes[1])
		return string(char), err
	}

	if len(runes) == 1 {
		return string(runes[0]), nil
	}
	return "", errors.New("not a char literal")
}

func (x *Lexer) DecodeAtom(atom string) (Token, error) {

	endColon := false
	n := len(atom)
	if atom[n-1] == ':' {
		endColon = true
		atom = atom[:n-1] // remove the colon
	}
	if atom == "&" {
		return x.Token(TokenSymbol, "&"), nil
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
	if atom == "NaN" || atom == "nan" {
		return x.Token(TokenFloat, "NaN"), nil
	}
	if DotSymbolRegex.MatchString(atom) {
		//Q("matched DotSymbolRegex '%v'", atom)
		return x.Token(TokenDotSymbol, atom), nil
	}
	if BuiltinOpRegex.MatchString(atom) {
		return x.Token(TokenSymbol, atom), nil
	}
	if atom == ":" {
		return x.Token(TokenSymbol, atom), nil
	} else if SymbolRegex.MatchString(atom) {
		////Q("matched symbol regex, atom='%v'", atom)
		if endColon {
			////Q("matched symbol regex with colon, atom[:n-1]='%v'", atom[:n-1])
			return x.Token(TokenSymbolColon, atom[:n-1]), nil
		}
		return x.Token(TokenSymbol, atom), nil
	}
	if CharRegex.MatchString(atom) {
		char, err := DecodeChar(atom)
		if err != nil {
			return x.EmptyToken(), err
		}
		return x.Token(TokenChar, char), nil
	}

	if endColon {
		return x.Token(TokenColonOperator, ":"), nil
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
	lexer.AppendToken(tok)
	return nil

}

// with block comments, we've got to tell
// the parser about them, so it can recognize
// when another line is needed to finish a
// block comment.
func (lexer *Lexer) dumpComment() {
	str := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.AppendToken(lexer.Token(TokenComment, str))
}

func (lexer *Lexer) dumpString() {
	str := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.AppendToken(lexer.Token(TokenString, str))
}

func (lexer *Lexer) dumpBacktickString() {
	str := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.AppendToken(lexer.Token(TokenBacktickString, str))
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
top:
	switch lexer.state {

	case LexerCommentBlock:
		//Q("lexer.state = LexerCommentBlock")
		if r == '\n' {
			_, err := lexer.buffer.WriteRune('\n')
			if err != nil {
				return err
			}
			lexer.dumpComment()
			// stay in LexerCommentBlock
			return nil
		}
		if r == '*' {
			lexer.state = LexerCommentBlockAsterisk
			return nil
		}
	case LexerCommentBlockAsterisk:
		//Q("lexer.state = LexerCommentBlockAsterisk")
		if r == '/' {
			_, err := lexer.buffer.WriteString("*/")
			if err != nil {
				return err
			}
			lexer.dumpComment()
			lexer.AppendToken(lexer.Token(TokenEndBlockComment, ""))
			lexer.state = LexerNormal
			return nil
		}
		_, err := lexer.buffer.WriteRune('*')
		if err != nil {
			return err
		}
		lexer.state = LexerCommentBlock
		goto writeRuneToBuffer

	case LexerFirstFwdSlash:
		//Q("lexer.state = LexerFirstFwdSlash")
		if r == '/' {
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.state = LexerCommentLine
			_, err = lexer.buffer.WriteString("//")
			return err
		}
		if r == '*' {
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			_, err = lexer.buffer.WriteString("/*")
			if err != nil {
				return err
			}
			lexer.state = LexerCommentBlock
			lexer.AppendToken(lexer.Token(TokenBeginBlockComment, ""))
			return nil
		}
		lexer.state = LexerBuiltinOperator
		lexer.prevrune = '/'
		err := lexer.dumpBuffer() // don't mix with token before the /
		if err != nil {
			return err
		}
		goto top // process the unknown rune r

	case LexerCommentLine:
		//Q("lexer.state = LexerCommentLine")
		if r == '\n' {
			//Q("lexer.state = LexerCommentLine sees end of line comment: '%s', going to LexerNormal", string(lexer.buffer.Bytes()))
			lexer.dumpComment()
			lexer.state = LexerNormal
			return nil
		}

	case LexerBacktickString:
		if r == '`' {
			lexer.dumpBacktickString()
			lexer.state = LexerNormal
			return nil
		}
		lexer.buffer.WriteRune(r)
		return nil

	case LexerStrLit:
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

	case LexerStrEscaped:
		char, err := EscapeChar(r)
		if err != nil {
			return err
		}
		lexer.buffer.WriteRune(char)
		lexer.state = LexerStrLit
		return nil

	case LexerUnquote:
		if r == '@' {
			lexer.AppendToken(lexer.Token(TokenTildeAt, ""))
		} else {
			lexer.AppendToken(lexer.Token(TokenTilde, ""))
			lexer.buffer.WriteRune(r)
		}
		lexer.state = LexerNormal
		return nil
	case LexerFreshAssignOrColon:
		lexer.state = LexerNormal

		// there was a ':' followed by either '=' or something other than '=',

		// so proceed to process the normal ':' actions.
		if lexer.buffer.Len() == 0 {
			if r == '=' {
				lexer.AppendToken(lexer.Token(TokenFreshAssign, ":="))
				return nil
			}
		}
		if r == '=' {
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.AppendToken(lexer.Token(TokenFreshAssign, ":="))
			return nil
		} else {
			// but still allow ':' to be a token terminator at the end of a word.
			_, err := lexer.buffer.WriteRune(':')
			if err != nil {
				return err
			}
			err = lexer.dumpBuffer()
			if err != nil {
				return err
			}
			goto top // process the unknown rune r in Normal mode
		}

	case LexerBuiltinOperator:
		//Q("in LexerBuiltinOperator")
		lexer.state = LexerNormal
		// three cases: negative number, one rune operator, two rune operator
		first := string(lexer.prevrune)
		atom := fmt.Sprintf("%c%c", lexer.prevrune, r)
		//Q("in LexerBuiltinOperator, first='%s', atom='%s'", first, atom)
		// are we a negative number -1 or -.1 rather than  ->, --, -= operator?
		if lexer.prevrune == '-' {
			if FloatRegex.MatchString(atom) || DecimalRegex.MatchString(atom) {
				//Q("'%s' is the beginning of a negative number", atom)
				_, err := lexer.buffer.WriteString(atom)
				if err != nil {
					return err
				}
				return nil
			} else {
				//Q("atom was not matched by FloatRegex: '%s'", atom)
			}
		}

		if BuiltinOpRegex.MatchString(atom) {
			//Q("2 rune atom in builtin op '%s', first='%s'", atom, first)
			// 2 rune op
			lexer.AppendToken(lexer.Token(TokenSymbol, atom))
			return nil
		}
		//Q("1 rune atom in builtin op '%s', first='%s'", atom, first)
		lexer.AppendToken(lexer.Token(TokenSymbol, first))
		goto top // still have to parse r in normal

	case LexerNormal:
		switch r {
		case '*':
			fallthrough
		case '+':
			fallthrough
		case '-':
			fallthrough
		case '<':
			fallthrough
		case '>':
			fallthrough
		case '=':
			fallthrough
		case '!':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.state = LexerBuiltinOperator
			lexer.prevrune = r
			return nil

		case '/':
			lexer.state = LexerFirstFwdSlash
			return nil

		case '`':
			if lexer.buffer.Len() > 0 {
				return errors.New("Unexpected backtick")
			}
			lexer.state = LexerBacktickString
			return nil

		case '"':
			if lexer.buffer.Len() > 0 {
				return errors.New("Unexpected quote")
			}
			lexer.state = LexerStrLit
			return nil

		case ';':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.AppendToken(lexer.Token(TokenSemicolon, ";"))
			return nil

		case ',':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.AppendToken(lexer.Token(TokenComma, ","))
			return nil

		// colon terminates a keyword symbol, e.g. in `mykey: "myvalue"`;
		// mykey is the symbol.
		// Exception: unless it is the := operator for fresh assigment.
		case ':':
			lexer.state = LexerFreshAssignOrColon
			// won't know if it is ':' alone or ':=' for sure
			// until we get the next rune
			return nil

		// likewise &
		case '&':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.AppendToken(lexer.Token(TokenSymbol, "&"))
			return nil

		case '%': // replaces ' as our quote shorthand
			if lexer.buffer.Len() > 0 {
				return errors.New("Unexpected % quote")
			}
			lexer.AppendToken(lexer.Token(TokenQuote, ""))
			return nil

		// caret '^' replaces backtick '`' as the start of a macro template, so
		// we can use `` as in Go for verbatim strings (strings with newlines, etc).
		case '^':
			if lexer.buffer.Len() > 0 {
				return errors.New("Unexpected ^ caret")
			}
			lexer.AppendToken(lexer.Token(TokenCaret, ""))
			return nil

		case '~':
			if lexer.buffer.Len() > 0 {
				return errors.New("Unexpected tilde")
			}
			lexer.state = LexerUnquote
			return nil

		case '(':
			fallthrough
		case ')':
			fallthrough
		case '[':
			fallthrough
		case ']':
			fallthrough
		case '{':
			fallthrough
		case '}':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			lexer.AppendToken(lexer.DecodeBrace(r))
			return nil
		case '\n':
			lexer.linenum++
			fallthrough
		case ' ':
			fallthrough
		case '\t':
			fallthrough
		case '\r':
			err := lexer.dumpBuffer()
			if err != nil {
				return err
			}
			return nil
		} // end switch r in LexerNormal state

	} // end switch lexer.state

writeRuneToBuffer:
	_, err := lexer.buffer.WriteRune(r)
	if err != nil {
		return err
	}
	return nil
}

func (lexer *Lexer) PeekNextToken() (tok Token, err error) {
	/*
		Q("\n in PeekNextToken()\n")
		defer func() {
			Q("\n done with PeekNextToken() -> returning tok='%v', err=%v. tok='%#v'. tok==EndTk? %v\n",
				tok, err, tok, tok == EndTk)
		}()
	*/
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
	/*
		Q("\n in GetNextToken()\n")
		defer func() {
			Q("\n done with GetNextToken() -> returning tok='%v', err=%v. lexer.buffer.String()='%s'\n",
				tok, err, lexer.buffer.String())
		}()
	*/
	tok, err = lexer.PeekNextToken()
	if err != nil || tok.typ == TokenEnd {
		return EndTk, err
	}
	lexer.tokens = lexer.tokens[1:]
	return tok, nil
}

func (lex *Lexer) PromoteNextStream() (ok bool) {
	/*
		Q("entering PromoteNextStream()!\n")
		defer func() {
			Q("done with PromoteNextStream, promoted=%v\n", ok)
		}()
	*/
	if len(lex.next) == 0 {
		return false
	}
	//Q("Promoting next stream!\n")
	lex.stream = lex.next[0]
	lex.next = lex.next[1:]
	return true
}

func (lex *Lexer) AddNextStream(s io.RuneScanner) {
	// in case we still have input available,
	// save new stuff for later
	lex.next = append(lex.next, s)

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
