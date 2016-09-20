// BEFORE: original straight-line code
func (parser *Parser) parseArray(depth int) (Sexp, error) {
...
            if tok.typ != TokenEnd {
                break getTok // got a token
            } else {
                return io.EOF // <<<<<<<<<<<<<   sad, done before finding ']'
            }
...
