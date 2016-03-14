package zygo

type Pratt struct {
	// vars for pratt parsing inside {}
	NextToken  Token
	CnodeStack []Sexp
	AccumTree  Sexp
	Cur        Token
	Next       Token
}

// Expression():
//
// From Douglas Crockford's article on Pratt parsing:
//   "Top Down Operator Precedence"
// http://javascript.crockford.com/tdop/tdop.html
//
// The heart of Pratt's technique is the expression
// function. It takes a right binding power that
// controls how aggressively it binds to tokens on its right.
// expression calls the nud method of the token.
//
// The nud is used to process literals, variables,
// and prefix operators.
//
// Then as long as the right binding
// power is less than the left binding power of the next
// token, the led method is invoked on the following
// token. The led is used to process infix and
// suffix operators. This process can be recursive
// because the nud and led
// methods can call expression.
//
// In pseudo Java script:
//
// var expression = function (rbp) {
//    var left;
//    var t = token;
//    advance();
//    left = t.nud();
//    while (rbp < token.lbp) {
//        t = token;
//        advance();
//        left = t.led(left);
//    }
//    return left;
// }
//
// Below is a working expression() parsing routine. Reproduces the
// original Pratt and Crockford formulation.
//
// AccumLeftTree holds the accumulated parse tree at any point in time.
//     "The parse Tree Up to this point, by consuming the tokens
//      to the left" would be a better-but-too-long name.
//
//  and AccumTree is the stuff to the left of the
//   current operator in the parse stream.
//
// data flows from NextToken -> cnode -> (possibly on the stack of t
//   recursive MunchLeft calls) -> into the AccumLeftTree tree.
//
//  better names: _left  -> AccumLeftTree (to be returned)
//                t      -> cnode; as it is the current token's qtree
//                           node to be processed. Once we grab this
//                           we always advance() past it
//                           before processing it, so that
//                           NextToken contains the
//                           following token.
//
//
//  meaning of rbp parameter: if you hit a token with
//  a  NextToken.Lbp < rbp, then don't bother calling MunchLeft,
//  stop and return what you have.
//
// better explanation:  rbp = a lower bound on descendant nodes
// precedence level, so we can
// guarantee the precenence-hill-climbing property (small precedence
// at the top) in the resulting parse tree.
//
/*
func (p *Pratt) Expression(rbp int, depth int) (Sexp, error) {
	lexer := parser.lexer
	var err error
	var tok Token

//		if tok.typ == TokenRCurly {
//			// pop off the }
//			_, _ = lexer.GetNextToken()
//			break
//		}


		cnode := p.NextToken

		//	if p.IsEOF() {
		//		return cnode
		//	}
		p.CnodeStack = append(p.NextToken, p.CnodeStack...)

		sx, err := p.Advance()
		if sx == SexpEnd || err != nil {
			// reset requested
			return p.AccumTree, err
		}

		if cnode.typ != TokenEnd {

			// munch_right() of atoms returns this/itself, in which
			// case: p.AccumTree = t; is the result.
			p.AccumTree = cnode.MunchRight(cnode)
			// DV(p.AccumTree->print("p.AccumTree: "));
		}

		for !p.IsEOF() && rbp < p.NextToken.Lbp {
			//assert(NextToken);

			cnode = p.NextToken
			p.CnodeStack[0] = p.NextToken
			//_cnode_stack.front() = NextToken;

			//DV(cnode->print("cnode:  "));

			p.Advance(0)
			if p.NextToken != nil {
				//p("NextToken = %v", NextToken)
			}

			// if cnode->munch_left() returns this/itself, then the net effect is: p.AccumTree = cnode;
			p.AccumTree = cnode.MunchLeft(cnode, p.AccumTree)

		}

		p.CnodeStack = p.CnodeStack[1:]
		//_cnode_stack.pop_front()
		return p.AccumTree
	return SexpNull, nil
}


// Advance sets p.NextToken
func (p *Parser) Advance() (Sexp, error) {
	lexer := p.lexer
	var err error
	var tok Token
getTok:
	for {
		tok, err = lexer.PeekNextToken()
		if err != nil {
			return SexpNull, err
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
	} // end for
	p.NextToken = tok
	return SexpNull, nil
}
*/
