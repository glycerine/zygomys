// AFTER: we call getMoreInput()
func (parser *Parser) parseArray(depth int) (Sexp, error) {
...

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
...
