// Start() commences the background infinite loop of parsing
func (p *Parser) Start() {
    go func() {
    defer close(p.Done)
    expressions := make([]Sexp, 0, SliceDefaultCap)
    for {
        expr, err := p.parseExpression(0)
        if err != nil || expr == SexpEnd {
            if err == ParserHaltRequested {
                return
            }
            err = p.getMoreInput(expressions, err) // SexpEnd means we need more input
            if err == ParserHaltRequested {
                return
            }
            expressions = make([]Sexp, 0, SliceDefaultCap)
        } else {
            expressions = append(expressions, expr)
        }
    }
}

