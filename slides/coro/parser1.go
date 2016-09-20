
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
