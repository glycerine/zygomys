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
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			right, err := pr.Expression(env, 50)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				plus, left, right,
			})
			P("MunchLeft for +: MakeList returned list: '%v'", list.SexpString())
			return list, nil
		},
	}

	sub := env.MakeSymbol("-")
	env.infixOps["-"] = &InfixOp{
		Sym: sub,
		Bp:  50,
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			right, err := pr.Expression(env, 50)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				sub, left, right,
			})
			P("MunchLeft for -: MakeList returned list: '%v'", list.SexpString())
			return list, nil
		},
	}
}

type RightMuncher func(env *Glisp, pr *Pratt) (Sexp, error)
type LeftMuncher func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error)

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
	if x == nil {
		P("x was nil")
	} else {
		P("x back is not nil and is of type %T/val = '%v', err = %v", x, x.SexpString(), err)
	}
	dup := env.Duplicate()
	ev, err := dup.EvalExpressions([]Sexp{x})
	if err != nil {
		return SexpNull, err
	}
	return ev, nil
}

type Pratt struct {
	NextToken  Sexp
	CnodeStack []Sexp
	AccumTree  Sexp
	//	Cur        Sexp

	Pos    int
	Stream []Sexp
}

func NewPratt(stream []Sexp) *Pratt {
	p := &Pratt{
		NextToken:  SexpNull,
		AccumTree:  SexpNull,
		CnodeStack: make([]Sexp, 0),
		Stream:     stream,
	}
	if len(stream) > 0 {
		p.NextToken = stream[0]
	}
	return p
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

func (p *Pratt) Expression(env *Glisp, rbp int) (ret Sexp, err error) {
	defer func() {
		P("Expression is returning Sexp ret = '%v'", ret.SexpString())
	}()
	cnode := p.NextToken
	if cnode != nil {
		P("top of Expression, rbp = %v, cnode = '%v'", rbp, cnode.SexpString())
	} else {
		P("top of Expression, rbp = %v, cnode is nil", rbp)
	}
	if p.IsEOF() {
		P("Expression sees IsEOF, returning cnode = %v", cnode.SexpString())
		return cnode, nil
	}
	p.CnodeStack = append([]Sexp{p.NextToken}, p.CnodeStack...)
	p.ShowCnodeStack()

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
		P("about to MunchRight on cnode = %v", cnode.SexpString())
		p.AccumTree, err = curOp.MunchRight(env, p)
		if err != nil {
			P("Expression(%v) MunchRight saw err = %v", rbp, err)
			return SexpNull, err
		}
		P("after MunchRight on cnode = %v, p.AccumTree = '%v'",
			cnode.SexpString(), p.AccumTree.SexpString())
	} else {
		// do this, or have the default MunchRight return itself.
		p.AccumTree = cnode
	}

	for !p.IsEOF() {
		nextLbp := env.LeftBindingPower(p.NextToken)
		P("nextLbp = %v", nextLbp)
		if rbp >= nextLbp {
			break
		}

		cnode = p.NextToken
		curOp = nil
		switch x := cnode.(type) {
		case *SexpSymbol:
			op, found := env.infixOps[x.name]
			if found {
				curOp = op
			}
		default:
			panic(fmt.Errorf("how to handle cnode type = %#v", cnode))
		}

		p.CnodeStack[0] = p.NextToken
		//_cnode_stack.front() = NextToken;

		P("in MunchLeft loop, before Advance, p.NextToken = %v",
			p.NextToken.SexpString())
		p.Advance()
		if p.Pos < len(p.Stream) {
			P("in MunchLeft loop, after Advance, p.NextToken = %v",
				p.NextToken.SexpString())
		}

		// if cnode->munch_left() returns this/itself, then
		// the net effect is: p.AccumTree = cnode;
		if curOp != nil && curOp.MunchLeft != nil {
			P("about to MunchLeft, cnode = %#v, p.AccumTree = %#v", cnode, p.AccumTree)
			p.AccumTree, err = curOp.MunchLeft(env, p, p.AccumTree)
			if err != nil {
				P("curOp.MunchLeft saw err = %v", err)
				return SexpNull, err
			}
		} else {
			// do this, or have the default MunchLeft return itself.
			p.AccumTree = cnode
		}

	}

	p.CnodeStack = p.CnodeStack[1:]
	//_cnode_stack.pop_front()
	P("at end of Expression(%v), returning p.AccumTree=%v, err=nil", rbp, p.AccumTree.SexpString())
	return p.AccumTree, nil
}

// Advance sets p.NextToken
func (p *Pratt) Advance() error {
	p.Pos++
	if p.Pos >= len(p.Stream) {
		return io.EOF
	}
	p.NextToken = p.Stream[p.Pos]
	P("end of Advance, p.NextToken = '%v'", p.NextToken.SexpString())
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

func (p *Pratt) ShowCnodeStack() {
	if len(p.CnodeStack) == 0 {
		P("CnodeStack is: empty")
		return
	}
	P("CnodeStack is:")
	for i := range p.CnodeStack {
		P("CnodeStack[%v] = %v", i, p.CnodeStack[i].SexpString())
	}
}
