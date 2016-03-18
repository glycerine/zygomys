package zygo

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
)

type Parser struct {
	lexer *Lexer
	env   *Glisp

	Done         chan bool
	reqStop      chan bool
	AddInput     chan io.RuneScanner
	ReqReset     chan io.RuneScanner
	ParsedOutput chan []ParserReply
	LastErr      chan error

	mut               sync.Mutex
	stopped           bool
	lastError         error
	sendMe            []ParserReply
	FlagSendNeedInput bool
}

type ParserReply struct {
	Expr []Sexp
	Err  error
}

func (env *Glisp) NewParser() *Parser {
	p := &Parser{
		env:          env,
		Done:         make(chan bool),
		reqStop:      make(chan bool),
		ReqReset:     make(chan io.RuneScanner),
		AddInput:     make(chan io.RuneScanner),
		ParsedOutput: make(chan []ParserReply),
		LastErr:      make(chan error),
		sendMe:       make([]ParserReply, 0, 1),
	}
	p.lexer = NewLexer(p)
	return p
}

func (p *Parser) Stop() error {
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

func (p *Parser) Start() {
	go p.InfiniteParsingLoop()
}

var ParserHaltRequested = fmt.Errorf("parser halt requested")
var ResetRequested = fmt.Errorf("parser reset requested")

func (p *Parser) InfiniteParsingLoop() {
	defer close(p.Done)
	expressions := make([]Sexp, 0, SliceDefaultCap)

	// maybe we already have input, be optimistic!
	// no need to call p.GetMoreInput() before staring
	// our loop.

	for {
		expr, err := p.ParseExpression(0)
		if err != nil || expr == SexpEnd {
			if err == ParserHaltRequested {
				return
			}
			err = p.GetMoreInput(expressions, err)
			if err == ParserHaltRequested {
				return
			}
			// GetMoreInput will have delivered what we gave them. Reset since we
			// don't own that memory any more.
			expressions = make([]Sexp, 0, SliceDefaultCap)
		} else {
			expressions = append(expressions, expr)
		}
	}
}

var ErrMoreInputNeeded = fmt.Errorf("parser needs more input")

// This function should *return* when it has more input
// for the parser/lexer, which will call it when they get wedged.
//
// Listeners on p.ParsedOutput should know the Convention: sending
// a length 0 []ParserReply on p.ParsedOutput channel means: we need more
// input! They should send some in on p.AddInput channel; or request
// and reset and simultaneously give us new input with p.ReqReset channel.
func (p *Parser) GetMoreInput(deliverThese []Sexp, errorToReport error) error {

	if len(deliverThese) == 0 && errorToReport == nil {
		p.FlagSendNeedInput = true
	} else {
		p.sendMe = append(p.sendMe,
			ParserReply{
				Expr: deliverThese,
				Err:  errorToReport,
			})
	}

	for {
		select {
		case <-p.reqStop:
			return ParserHaltRequested
		case input := <-p.AddInput:
			p.lexer.AddNextStream(input)
			p.FlagSendNeedInput = false
			return nil
		case input := <-p.ReqReset:
			p.lexer.Reset()
			p.lexer.AddNextStream(input)
			p.FlagSendNeedInput = false
			return ResetRequested
		case p.HaveStuffToSend() <- p.sendMe:
			p.sendMe = make([]ParserReply, 0, 1)
			p.FlagSendNeedInput = false
		case p.LastErr <- p.lastError:
			p.lastError = nil
		}
	}
}

func (p *Parser) HaveStuffToSend() chan []ParserReply {
	if len(p.sendMe) > 0 || p.FlagSendNeedInput {
		return p.ParsedOutput
	}
	return nil
}

func (p *Parser) Reset() {
	select {
	case p.ReqReset <- nil:
	case <-p.reqStop:
	}
}

func (p *Parser) NewInput(s io.RuneScanner) {
	select {
	case p.AddInput <- s:
	case <-p.reqStop:
	}
}

func (p *Parser) ResetAddNewInput(s io.RuneScanner) {
	select {
	case p.ReqReset <- s:
	case <-p.reqStop:
	}
}

var UnexpectedEnd error = errors.New("Unexpected end of input")

const SliceDefaultCap = 10

func (parser *Parser) ParseList(depth int) (sx Sexp, err error) {
	lexer := parser.lexer
	var tok Token

tokFilled:
	for {
		tok, err = lexer.PeekNextToken()
		//Q("\n ParseList(depth=%d) got lexer.PeekNextToken() -> tok='%v' err='%v'\n", depth, tok, err)
		if err != nil {
			return SexpNull, err
		}
		if tok.typ != TokenEnd {
			break tokFilled
		}
		// instead of returning UnexpectedEnd, we:
		err = parser.GetMoreInput(nil, ErrMoreInputNeeded)
		//Q("\n ParseList(depth=%d) got back from parser.GetMoreInput(): '%v'\n", depth, err)
		switch err {
		case ParserHaltRequested:
			return SexpNull, err
		case ResetRequested:
			return SexpEnd, err
		}
		// have to still fill tok, so
		// loop to the top to PeekNextToken
	}

	if tok.typ == TokenRParen {
		_, _ = lexer.GetNextToken()
		return SexpNull, nil
	}

	var start = &SexpPair{}

	expr, err := parser.ParseExpression(depth + 1)
	if err != nil {
		return SexpNull, err
	}

	start.Head = expr

	tok, err = lexer.PeekNextToken()
	if err != nil {
		return SexpNull, err
	}

	// backslash '\' replaces dot '.' in zygomys
	if tok.typ == TokenBackslash {
		// eat up the backslash
		_, _ = lexer.GetNextToken()
		expr, err = parser.ParseExpression(depth + 1)
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
		start.Tail = expr
		return start, nil
	}

	expr, err = parser.ParseList(depth + 1)
	if err != nil {
		return start, err
	}
	start.Tail = expr

	return start, nil
}

func (parser *Parser) ParseArray(depth int) (Sexp, error) {
	lexer := parser.lexer
	arr := make([]Sexp, 0, SliceDefaultCap)

	var tok Token
	var err error
	for {
	getTok:
		for {
			tok, err = lexer.PeekNextToken()
			if err != nil {
				return SexpEnd, err
			}

			if tok.typ != TokenEnd {
				break getTok
			} else {
				//instead of return SexpEnd, UnexpectedEnd
				// we ask for more, and then loop
				err = parser.GetMoreInput(nil, ErrMoreInputNeeded)
				switch err {
				case ParserHaltRequested:
					return SexpNull, err
				case ResetRequested:
					return SexpEnd, err
				}
			}
		}

		if tok.typ == TokenRSquare {
			// pop off the ]
			_, _ = lexer.GetNextToken()
			break
		}

		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	return &SexpArray{Val: arr}, nil
}

func (parser *Parser) ParseExpression(depth int) (res Sexp, err error) {
	defer func() {
		if res != nil {
			//P("returning from ParseExpression at depth=%v with res='%s'\n", depth, res.SexpString())
		} else {
			//P("returning from ParseExpression at depth=%v, res = nil", depth)
		}
	}()

	lexer := parser.lexer
	env := parser.env

getAnother:
	tok, err := lexer.GetNextToken()
	if err != nil {
		return SexpEnd, err
	}

	switch tok.typ {
	case TokenLParen:
		exp, err := parser.ParseList(depth + 1)
		return exp, err
	case TokenLSquare:
		exp, err := parser.ParseArray(depth + 1)
		return exp, err
	case TokenLCurly:
		exp, err := parser.ParseInfix(depth + 1)
		return exp, err
	case TokenQuote:
		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("quote"), expr}), nil
	case TokenCaret:
		// '^' is now our syntax-quote symbol, not TokenBacktick, to allow go-style `string literals`.
		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("syntaxQuote"), expr}), nil
	case TokenTilde:
		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("unquote"), expr}), nil
	case TokenTildeAt:
		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("unquote-splicing"), expr}), nil
	case TokenFreshAssign:
		return env.MakeSymbol(tok.str), nil
	case TokenColonOperator:
		return env.MakeSymbol(tok.str), nil
	case TokenDollar:
		return env.MakeSymbol(tok.str), nil
	case TokenBool:
		return &SexpBool{Val: tok.str == "true"}, nil
	case TokenDecimal:
		i, err := strconv.ParseInt(tok.str, 10, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return &SexpInt{Val: i}, nil
	case TokenHex:
		i, err := strconv.ParseInt(tok.str, 16, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return &SexpInt{Val: i}, nil
	case TokenOct:
		i, err := strconv.ParseInt(tok.str, 8, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return &SexpInt{Val: i}, nil
	case TokenBinary:
		i, err := strconv.ParseInt(tok.str, 2, SexpIntSize)
		if err != nil {
			return SexpNull, err
		}
		return &SexpInt{Val: i}, nil
	case TokenChar:
		return &SexpChar{Val: rune(tok.str[0])}, nil
	case TokenString:
		return &SexpStr{S: tok.str}, nil
	case TokenBacktickString:
		return &SexpStr{S: tok.str, backtick: true}, nil
	case TokenFloat:
		f, err := strconv.ParseFloat(tok.str, SexpFloatSize)
		if err != nil {
			return SexpNull, err
		}
		return &SexpFloat{Val: f}, nil
	case TokenEnd:
		return SexpEnd, nil
	case TokenSymbol:
		return env.MakeSymbol(tok.str), nil
	case TokenSymbolColon:
		sym := env.MakeSymbol(tok.str)
		sym.colonTail = true
		return sym, nil
	case TokenDot:
		sym := env.MakeSymbol(tok.str)
		sym.isDot = true
		return sym, nil
	case TokenDotSymbol:
		sym := env.MakeSymbol(tok.str)
		sym.isDot = true
		return sym, nil
	case TokenComment:
		//P("parser making SexpComment from '%s'", tok.str)
		return &SexpComment{Comment: tok.str}, nil
		// parser skips comments
		//goto getAnother
	case TokenBeginBlockComment:
		// parser skips comments
		return parser.ParseBlockComment(&tok)
		//parser.ParseBlockComment(&tok)
		//goto getAnother
	case TokenComma:
		goto getAnother
	case TokenSemicolon:
		goto getAnother // not the right thing, still TODO to figure out what is.
	}
	return SexpNull, fmt.Errorf("Invalid syntax, don't know what to do with %v '%v'", tok.typ, tok)
}

// ParseTokens is the main service the Parser provides.
// Currently returns first error encountered, ignoring
// any expressions after that.
func (p *Parser) ParseTokens() ([]Sexp, error) {
	select {
	case out := <-p.ParsedOutput:
		Q("ParseTokens got p.ParsedOutput out: '%#v'", out)
		r := make([]Sexp, 0)
		for _, k := range out {
			r = append(r, k.Expr...)
			Q("\n ParseTokens k.Expr = '%v'\n\n", (&SexpArray{Val: k.Expr}).SexpString(0))
			if k.Err != nil {
				return r, k.Err
			}
		}
		return r, nil
	case <-p.reqStop:
		return nil, ErrShuttingDown
	}
}

var ErrShuttingDown error = fmt.Errorf("lexer shutting down")

func (parser *Parser) ParseBlockComment(start *Token) (sx Sexp, err error) {
	defer func() {
		if sx != nil {
			//P("returning from ParseBlockComment with sx ='%v', err='%v'",
			//	sx.SexpString(), err)
		}
	}()
	lexer := parser.lexer
	var tok Token
	var block = &SexpComment{Block: true, Comment: start.str}

	for {
	tokFilled:
		for {
			tok, err = lexer.PeekNextToken()
			if err != nil {
				return SexpNull, err
			}
			if tok.typ != TokenEnd {
				break tokFilled
			}
			err = parser.GetMoreInput(nil, ErrMoreInputNeeded)
			switch err {
			case ParserHaltRequested:
				return SexpNull, err
			case ResetRequested:
				return SexpEnd, err
			}
			// have to still fill tok, so
			// loop to the top to PeekNextToken
		}

		// consume it

		//cons, err := lexer.GetNextToken()
		_, err := lexer.GetNextToken()
		if err != nil {
			return nil, err
		}
		//P("parse block comment is consuming '%v'", cons)

		switch tok.typ {
		case TokenEndBlockComment:
			block.Comment += tok.str
			return block, nil
		case TokenComment:
			block.Comment += tok.str
		default:
			panic("internal error: inside a block comment, we should only see TokenComment and TokenEndBlockComment tokens")
		}
	}
	//return block, nil
}

func (parser *Parser) ParseInfix(depth int) (Sexp, error) {
	lexer := parser.lexer
	arr := make([]Sexp, 0, SliceDefaultCap)
	var err error
	var tok Token
	for {
	getTok:
		for {
			tok, err = lexer.PeekNextToken()
			if err != nil {
				return SexpEnd, err
			}

			if tok.typ != TokenEnd {
				break getTok
			} else {
				//instead of return SexpEnd, UnexpectedEnd
				// we ask for more, and then loop
				err = parser.GetMoreInput(nil, ErrMoreInputNeeded)
				switch err {
				case ParserHaltRequested:
					return SexpNull, err
				case ResetRequested:
					return SexpEnd, err
				}
			}
		}

		if tok.typ == TokenRCurly {
			// pop off the }
			_, _ = lexer.GetNextToken()
			break
		}

		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	var list SexpPair
	list.Head = parser.env.MakeSymbol("infix")
	list.Tail = SexpNull
	if len(arr) > 0 {
		list.Tail = Cons(&SexpArray{Val: arr, Infix: true}, SexpNull)
	}
	return &list, nil

	//return &SexpArray{Val: arr, Infix: true}, nil
}
