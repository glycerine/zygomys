// AFTER in context
func (parser *Parser) parseArray(depth int) (Sexp, error) {
    for { // get the next token, then break
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
                break getTok // got a token
            } else {
                // we ask for more, and then loop
                err = parser.getMoreInput(nil, ErrMoreInputNeeded) // <<<=== key change
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
