package zygo

import (
	"fmt"
	"io"
)

type InfixOp struct {
	Sym        *SexpSymbol
	Bp         int
	MunchRight RightMuncher
	MunchLeft  LeftMuncher
}

func (env *Glisp) InitInfixOps() {
	plus := env.MakeSymbol("+")
	env.infixOps["+"] = &InfixOp{
		Sym: plus,
		Bp:  50,
		MunchLeft: func(ths, left Sexp) Sexp {
			return MakeList([]Sexp{
				plus, ths, left,
			})
		},
	}
}

type RightMuncher func(ths Sexp) Sexp
type LeftMuncher func(ths Sexp, left Sexp) Sexp

func InfixBuilder(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		// let {} mean nil
		return SexpNull, nil
	}
	var arr *SexpArray
	switch v := args[0].(type) {
	case *SexpArray:
		arr = v
	default:
		return SexpNull, fmt.Errorf("InfixBuilder must receive an SexpArray")
	}
	P("InfixBuilder, arr = ")
	for i := range arr.Val {
		P("arr[%v] = %v of type %T", i, arr.Val[i].SexpString(), arr.Val[i])
	}
	pr := NewPratt(arr.Val)
	x, err := pr.Expression(env, 0)
	P("x = %v, err = %v", x, err)
	return arr, nil
}

type Pratt struct {
	NextToken  Sexp
	CnodeStack []Sexp
	AccumTree  Sexp
	Cur        Sexp

	Pos    int
	Stream []Sexp
}

func NewPratt(stream []Sexp) *Pratt {
	return &Pratt{
		CnodeStack: make([]Sexp, 0),
		Stream:     stream,
	}
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

func (p *Pratt) Expression(env *Glisp, rbp int) (Sexp, error) {
	cnode := p.NextToken

	if p.IsEOF() {
		return cnode, nil
	}
	p.CnodeStack = append([]Sexp{p.NextToken}, p.CnodeStack...)

	p.Advance()
	/*
		err := p.Advance()
			switch err {
			case io.EOF:
				return p.AccumTree, nil
			default:
				return p.AccumTree, err
			case nil:
			}
	*/

	var curOp *InfixOp
	switch x := cnode.(type) {
	case *SexpSymbol:
		op, found := env.infixOps[x.name]
		if found {
			curOp = op
		}
	}

	if curOp != nil && curOp.MunchRight != nil {
		// munch_right() of atoms returns this/itself, in which
		// case: p.AccumTree = t; is the result.
		p.AccumTree = curOp.MunchRight(cnode)
	}

	for !p.IsEOF() && rbp < env.LeftBindingPower(p.NextToken) {
		//assert(NextToken);

		cnode = p.NextToken
		curOp = nil
		switch x := cnode.(type) {
		case *SexpSymbol:
			op, found := env.infixOps[x.name]
			if found {
				curOp = op
			}
		}

		p.CnodeStack[0] = p.NextToken
		//_cnode_stack.front() = NextToken;

		p.Advance()
		if p.Pos < len(p.Stream) {
			P("p.NextToken = %v", p.NextToken)
		}

		// if cnode->munch_left() returns this/itself, then
		// the net effect is: p.AccumTree = cnode;
		if curOp != nil && curOp.MunchLeft != nil {
			p.AccumTree = curOp.MunchLeft(cnode, p.AccumTree)
		}
	}

	p.CnodeStack = p.CnodeStack[1:]
	//_cnode_stack.pop_front()
	return p.AccumTree, nil
}

// Advance sets p.NextToken
func (p *Pratt) Advance() error {
	p.Pos++
	if p.Pos >= len(p.Stream) {
		return io.EOF
	}
	p.NextToken = p.Stream[p.Pos]
	return nil
}

func (p *Pratt) IsEOF() bool {
	if p.Pos >= len(p.Stream) {
		return true
	}
	return false
}

func (env *Glisp) LeftBindingPower(sx Sexp) int {
	switch x := sx.(type) {
	case *SexpInt:
		return 0
	case *SexpSymbol:
		op, found := env.infixOps[x.name]
		if found {
			return op.Bp
		}
		return 0
	}
	return 0
}
