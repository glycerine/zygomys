package main

// a simplified version of github.com/glycerine/zygomys/repl/parser.go

type Parser struct {
	Done    chan bool
	reqStop chan bool

	// client adds input here, typically by calling Parser.NewInput().
	// io.RuneScanner supports ReadRune() and UnreadRune().
	AddInput chan io.RuneScanner

	// client resets old parse and adds new input here,
	// typically by calling Parser.ResetAddNewInput()
	ReqReset chan io.RuneScanner

	// client obtains output here, typically by calling
	// Parser.ParseTokens()
	ParsedOutput chan []ParserReply

	mut               sync.Mutex
	stopped           bool
	sendMe            []ParserReply
	FlagSendNeedInput bool
}

type ParserReply struct {
	Expr []Sexp
	Err  error
}

// NewParser creates a new stopped Parser. Call Start()
// on it before using it.
func (env *Glisp) NewParser() *Parser {
	p := &Parser{
		env:     env,
		Done:    make(chan bool),
		reqStop: make(chan bool),

		ReqReset:     make(chan io.RuneScanner),
		AddInput:     make(chan io.RuneScanner),
		ParsedOutput: make(chan []ParserReply),
		sendMe:       make([]ParserReply, 0, 1),
	}
	p.lexer = NewLexer(p)
	return p
}

// ParseTokens is the main service the Parser provides.
// Currently returns first error encountered, ignoring
// any expressions after that.
func (p *Parser) ParseTokens() ([]Sexp, error) {
	select {
	case out := <-p.ParsedOutput:
		// out is type []ParserReply
		r := make([]Sexp, 0)
		for _, k := range out {
			r = append(r, k.Expr...)
			if k.Err != nil {
				return r, k.Err
			}
		}
		return r, nil
	case <-p.reqStop:
		return nil, ErrShuttingDown
	}
}

// NewInput is the principal API function to
// supply parser with addition textual
// input lines
func (p *Parser) NewInput(s io.RuneScanner) {
	select {
	case p.AddInput <- s:
	case <-p.reqStop:
	}
}

// ResetAddNewInput is the principal API function to
// tell the parser to forget everything it has stored,
// reset, and take as new input the scanner s.
func (p *Parser) ResetAddNewInput(s io.RuneScanner) {
	select {
	case p.ReqReset <- s:
	case <-p.reqStop:
	}
}

var ParserHaltRequested = fmt.Errorf("parser halt requested")
var ResetRequested = fmt.Errorf("parser reset requested")

// Stop gracefully shutsdown the parser and its background goroutine.
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

// Start() commences the background parse loop goroutine.
func (p *Parser) Start() {
	go p.infiniteParsingLoop()
}

var ParserHaltRequested = fmt.Errorf("parser halt requested")
var ResetRequested = fmt.Errorf("parser reset requested")

func (p *Parser) infiniteParsingLoop() {
	defer close(p.Done)
	expressions := make([]Sexp, 0, SliceDefaultCap)

	// maybe we already have input, be optimistic!
	// no need to call p.getMoreInput() before staring
	// our loop. The client may have already loaded
	// some text or a stream that already
	// has input ready for us.

	for {
		expr, err := p.parseExpression(0)
		if err != nil || expr == SexpEnd {
			if err == ParserHaltRequested {
				return
			}
			// expr == SexpEnd means that parserExpression
			// couldn't read another token, so a call to
			// getMoreInput() is required.

			// provide accumulated expressions
			// back to the client here
			err = p.getMoreInput(expressions, err)
			if err == ParserHaltRequested {
				return
			}

			// getMoreInput() will have delivered
			// expressions to the client. Reset expressions since we
			// don't own that memory any more.
			expressions = make([]Sexp, 0, SliceDefaultCap)
		} else {
			// INVAR: err == nil && expr is not SexpEnd
			expressions = append(expressions, expr)
		}
	}
}

var ErrMoreInputNeeded = fmt.Errorf("parser needs more input")

// getMoreInput is called by the Parser routines mid-parse, if
// need be, to obtain the next line/rune of input.
//
// getMoreInput() is used by Parser.ParseList(), Parser.ParseArray(),
// Parser.ParseBlockComment(), and Parser.ParseInfix().
//
// getMoreInput() is also used by Parser.infiniteParsingLoop() which
// is the main driver behind parsing.
//
// This function should *return* when it has more input
// for the parser/lexer, which will call it when they get wedged.
//
// Listeners on p.ParsedOutput should know the Convention: sending
// a length 0 []ParserReply on p.ParsedOutput channel means: we need more
// input! They should send some in on p.AddInput channel; or request
// a reset and simultaneously give us new input with p.ReqReset channel.
func (p *Parser) getMoreInput(deliverThese []Sexp, errorToReport error) error {

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
			// that was a conditional send, because
			// HaveStuffToSend() will return us a
			// nil channel if there's nothing ready.
			p.sendMe = make([]ParserReply, 0, 1)
			p.FlagSendNeedInput = false
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

var UnexpectedEnd error = errors.New("Unexpected end of input")

// ParserExpression is an internal Parser routine -  the
// main one for parsing expressions
func (parser *Parser) parseExpression(depth int) (res Sexp, err error) {

	//getAnother:
	tok, err := parser.lexer.getNextToken()
	if err != nil {
		return SexpEnd, err
	}

	switch tok.typ {
	case TokenLParen:
		exp, err := parser.parseList(depth + 1)
		return exp, err
	case TokenLSquare:
		exp, err := parser.parseArray(depth + 1)
		return exp, err
	case TokenLCurly:
		exp, err := parser.parseInfix(depth + 1)
		return exp, err
	case TokenQuote:
		expr, err := parser.parseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{parser.env.MakeSymbol("quote"), expr}), nil
	case TokenCaret:
		//...
	}
}

func (parser *Parser) parseArray(depth int) (Sexp, error) {
	arr := make([]Sexp, 0, SliceDefaultCap)

	var tok Token
	var err error
	for {
	getTok:
		for {
			tok, err = parser.lexer.peekNextToken()
			if err != nil {
				return SexpEnd, err
			}

			if tok.typ == TokenComma {
				// pop off the ,
				_, _ = parser.lexer.getNextToken()
				continue getTok
			}

			if tok.typ != TokenEnd {
				break getTok
			} else {
				// we ask for more, and then loop
				err = parser.getMoreInput(nil, ErrMoreInputNeeded)
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
			_, _ = parser.lexer.getNextToken()
			break
		}

		expr, err := parser.parseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	return &SexpArray{Val: arr, Env: parser.env}, nil
}

func (parser *Parser) parseList(depth int) (sx Sexp, err error) {
	var tok Token

tokFilled:
	for {
		// if lexer runs out of tokens it will
		// return EndTk = Token{typ: TokenEnd}.
		//
		tok, err = parser.lexer.peekNextToken()
		if err != nil {
			return SexpNull, err
		}
		if tok.typ != TokenEnd {
			break tokFilled
		}
		// instead of returning UnexpectedEnd, we:
		err = parser.getMoreInput(nil, ErrMoreInputNeeded)

		switch err {
		case ParserHaltRequested:
			return SexpNull, err
		case ResetRequested:
			return SexpEnd, err
		}
		// have to still fill tok, so
		// loop to the top to peekNextToken
	}

	if tok.typ == TokenRParen {
		_, _ = parser.lexer.getNextToken()
		return SexpNull, nil
	}

	var start = &SexpPair{}

	expr, err := parser.parseExpression(depth + 1)
	if err != nil {
		return SexpNull, err
	}

	start.Head = expr

	tok, err = parser.lexer.peekNextToken()
	if err != nil {
		return SexpNull, err
	}

	// backslash '\' replaces dot '.' in zygomys
	if tok.typ == TokenBackslash {
		// eat up the backslash
		_, _ = parser.lexer.getNextToken()
		expr, err = parser.parseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}

		// eat up the end paren
		tok, err = parser.lexer.getNextToken()
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

	expr, err = parser.parseList(depth + 1)
	if err != nil {
		return start, err
	}
	start.Tail = expr

	return start, nil
}

func (parser *Parser) ParseBlockComment(start *Token) (sx Sexp, err error) {
	defer func() {
		if sx != nil {
			//Q("returning from ParseBlockComment with sx ='%v', err='%v'",
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
		//Q("parse block comment is consuming '%v'", cons)

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

		Q("debug: ParseInfix(depth=%v) calling ParseExpression", depth)
		expr, err := parser.ParseExpression(depth + 1)
		if err != nil {
			return SexpNull, err
		}
		Q("debug2: ParseInfix(depth=%v) appending expr = '%v'", depth, expr.SexpString(nil))

		arr = append(arr, expr)
	}

	var list SexpPair
	list.Head = parser.env.MakeSymbol("infix")
	list.Tail = SexpNull
	if len(arr) > 0 {
		list.Tail = Cons(&SexpArray{Val: arr, Infix: true, Env: parser.env}, SexpNull)
	}
	return &list, nil
	//return &SexpArray{Val: arr, Infix: true, Env: env}, nil
}
