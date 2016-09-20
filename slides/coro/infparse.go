func (p *Parser) infiniteParsingLoop() {
	defer close(p.Done)
	expressions := make([]Sexp, 0, SliceDefaultCap)
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
