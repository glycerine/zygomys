// (original straight-line code:) parseArray handles `[2, 4, 5, "six"]` arrays of expressions
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
    	        return nil, io.EOF // <<<<<<<<<<<<<   sad, done before finding ']'
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
