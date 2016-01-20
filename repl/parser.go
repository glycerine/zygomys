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

	Ready        chan bool
	Done         chan bool
	reqStop      chan bool
	AddInput     chan io.RuneScanner
	ReqReset     chan io.RuneScanner
	ParsedOutput chan []ParserReply
	LastErr      chan error

	mut       sync.Mutex
	stopped   bool
	readySexp []Sexp
	lastError error
	sendMe    []ParserReply
}

type ParserReply struct {
	Expr []Sexp
	Err  error
}

func (env *Glisp) NewParser() *Parser {
	p := &Parser{
		env:          env,
		Ready:        make(chan bool),
		Done:         make(chan bool),
		reqStop:      make(chan bool),
		ReqReset:     make(chan io.RuneScanner),
		AddInput:     make(chan io.RuneScanner),
		ParsedOutput: make(chan []ParserReply, 1000),
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
	go p.InifiniteParsingLoop()
	<-p.Ready
}

var ParserHaltRequested = fmt.Errorf("parser halt requested")
var ResetRequested = fmt.Errorf("parser reset requested")

func (p *Parser) InifiniteParsingLoop() {
	P("InifiniteParsingLoop() started\n")
	expressions := make([]Sexp, 0, SliceDefaultCap)

	err := p.GetMoreInput(nil, nil)
	P("InifiniteParsingLoop() done with *initial* GetMoreInput, err = %v\n", err)
	switch err {
	case ParserHaltRequested:
		return
	case ResetRequested:
		// nothing else to do here
	}

	for {
		expr, err := p.ParseExpression(0)
		P("InifiniteParsingLoop() done with ParseExprssion with err = %v\n", err)
		if err != nil {
			if err == ParserHaltRequested {
				return
			}
			err = p.GetMoreInput(expressions, err)
			P("InifiniteParsingLoop() done GetMoreInput, err = %v\n", err)
			switch err {
			case ParserHaltRequested:
				return
			case ResetRequested:
				// nothing else to do here
			}
			expressions = make([]Sexp, 0, SliceDefaultCap)
		}
		expressions = append(expressions, expr)
	}
}

// this function should *return* when it has more input
func (p *Parser) GetMoreInput(deliverThese []Sexp, errorToReport error) error {
	if len(deliverThese) == 0 && errorToReport == nil {
		// nothing to report, leave sendMe alone.
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
			P("parser reqStop called!\n")
			return ParserHaltRequested
		case input := <-p.AddInput:
			P("Parser AddInput called!\n")
			p.lexer.AddNextStream(input)
			return nil
		case input := <-p.ReqReset:
			P("p.ReqReset called with input %p\n", input)
			p.lexer.Reset()
			p.lexer.AddNextStream(input)
			return ResetRequested
		case p.HaveStuffToSend() <- p.sendMe:
			P("Parser sent %v p.sendMe on ParsedOutput:\n", len(p.sendMe))
			for i := range p.sendMe {
				P("p.sendMe[%d]=%v\n", i, p.sendMe[i])
			}
			p.sendMe = make([]ParserReply, 0, 1)

		case p.LastErr <- p.lastError:
			P("Parser sent lastError %s on LastErr channel\n", p.lastError)
			p.lastError = nil

		}
	}
}

func (p *Parser) HaveStuffToSend() chan []ParserReply {
	if len(p.sendMe) > 0 {
		return p.ParsedOutput
	}
	return nil
}

func (p *Parser) finish() {
	close(p.Done)
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
	<-p.Ready
	select {
	case p.ReqReset <- s:
	case <-p.reqStop:
	}
}

var UnexpectedEnd error = errors.New("Unexpected end of input")

const SliceDefaultCap = 10

func (parser *Parser) ParseList(depth int) (Sexp, error) {
	lexer := parser.lexer
	var tok Token
	var err error
tokFilled:
	for {
		tok, err = lexer.PeekNextToken()
		if err != nil {
			return SexpNull, err
		}
		if tok.typ != TokenEnd {
			break tokFilled
		}
		// instead of retutning UnexpectedEnd, we:
		err = parser.GetMoreInput(nil, nil)
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

	var start SexpPair

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
				err = parser.GetMoreInput(nil, nil)
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

	return SexpArray(arr), nil
}

func (parser *Parser) ParseHash(depth int) (Sexp, error) {
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
				err = parser.GetMoreInput(nil, nil)
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
	list.Head = parser.env.MakeSymbol("hash")
	list.Tail = MakeList(arr)

	return list, nil
}

func (parser *Parser) ParseExpression(depth int) (res Sexp, err error) {
	lexer := parser.lexer
	env := parser.env

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
		exp, err := parser.ParseHash(depth + 1)
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
		return MakeList([]Sexp{env.MakeSymbol("syntax-quote"), expr}), nil
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
	case TokenSymbol:
		return env.MakeSymbol(tok.str), nil
	case TokenColonOperator:
		return env.MakeSymbol(tok.str), nil
	case TokenDollar:
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
	case TokenOct:
		i, err := strconv.ParseInt(tok.str, 8, SexpIntSize)
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
	return SexpNull, errors.New(fmt.Sprint("Invalid syntax, didn't know what to do with ", tok.typ, " ", tok))
}

// private main service routine starts here.
func (parser *Parser) parseTokens() ([]Sexp, error) {
	P("parseTokens called!\n")
	expressions := make([]Sexp, 0, SliceDefaultCap)

	for {
		expr, err := parser.ParseExpression(0)
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

// ParseTokens is the main service the Parser provides.
func (p *Parser) ParseTokens() ([]Sexp, error) {
	P("ParseTokens called!\n")
	select {
	case out := <-p.ParsedOutput:
		r := make([]Sexp, 0)
		for _, k := range out {
			r = append(r, k.Expr...)
			if k.Err != nil {
				V("\n ParseTokens() sees err %v\n", k.Err)
				return r, k.Err
			}
		}
		return r, nil
	case <-p.reqStop:
		return nil, ErrShuttingDown
	}
}

var ErrShuttingDown error = fmt.Errorf("lexer shutting down")
