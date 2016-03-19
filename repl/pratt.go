package zygo

import (
	"fmt"
	"io"
)

// Pratt parsing. see http://javascript.crockford.com/tdop/tdop.html
// Also nice writeup: http://journal.stuffwithstuff.com/2011/03/19/pratt-parsers-expression-parsing-made-easy/

// precedence levels (smaller == lower priority,
//    so smaller => goes towards top of tree)
//
//  Borrowing from the tdop.html precedence list mostly:
//
//  0  non-binding operators like ;
// 10  assignment operators like = :=
// 20  ?
// 30  or and
// 40  relational operators like ==
// 50  + -
// 60  * /
// 65  **
// 70  unary operators like 'not'
// 80  . [ (
//

// InfixOp lets us attach led (MunchLeft) and nud (MunchRight)
// Pratt parsing methods, along with a binding power, to a symbol.
type InfixOp struct {
	Sym        *SexpSymbol
	Bp         int          // binding power, aka precedence level.
	MunchRight RightMuncher // aka nud
	MunchLeft  LeftMuncher  // aka led
}

// Infix creates a new infix operator
func (env *Glisp) Infix(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			right, err := pr.Expression(env, bp)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				oper, left, right,
			})
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

func (env *Glisp) InfixF(op string, bp int, f func(env *Glisp, op string, bp int) *InfixOp) *InfixOp {
	return f(env, op, bp)
}

// Infix creates a new (right-associative) short-circuiting
// infix operator, used for `and` and `or` in infix processing.
func (env *Glisp) Infixr(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			right, err := pr.Expression(env, bp-1)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				oper, left, right,
			})
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

// Prefix creates a new prefix operator, like `not`, for
// infix processing.
func (env *Glisp) Prefix(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchRight: func(env *Glisp, pr *Pratt) (Sexp, error) {
			right, err := pr.Expression(env, bp)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				oper, right,
			})
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

// Assignment creates a new assignment operator for infix
// processing.
func (env *Glisp) Assignment(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	operSet := env.MakeSymbol("set")
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			// TODO: check that left is okay as an LVALUE

			right, err := pr.Expression(env, bp-1)
			if err != nil {
				return SexpNull, err
			}
			if op == "=" || op == ":=" {
				oper = operSet
			}

			list := MakeList([]Sexp{
				oper, left, right,
			})
			Q("assignment returning list: '%v'", list.SexpString(0))
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

// PostfixAssign creates a new postfix assignment operator for infix
// processing.
func (env *Glisp) PostfixAssign(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
			// TODO: check that left is okay as an LVALUE
			list := MakeList([]Sexp{
				oper, left,
			})
			Q("postfix assignment returning list: '%v'", list.SexpString(0))
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

func arrayOpMunchLeft(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
	oper := env.MakeSymbol("arrayidx")
	Q("pr.NextToken = '%v', left = %#v", pr.NextToken.SexpString(0), left)
	if len(pr.CnodeStack) > 0 {
		Q("pr.CnodeStack[0] = '%v'", pr.CnodeStack[0])
	}
	//right, err := pr.Expression(env, 0)
	right := pr.NextToken
	Q("right = %#v", right)
	list := MakeList([]Sexp{
		oper, left, pr.CnodeStack[0],
	})
	return list, nil
}

var arrayOp *InfixOp

// InitInfixOps establishes the env.infixOps definitions
// required for infix parsing using the Pratt parser.
func (env *Glisp) InitInfixOps() {
	env.Infix("+", 50)
	env.Infix("-", 50)
	env.Infix("*", 60)
	env.Infix("/", 60)
	env.Infix("mod", 60)
	env.Infixr("**", 65)
	env.Infixr("and", 30)
	env.Infixr("or", 30)
	env.Prefix("not", 70)
	env.Assignment("=", 10)
	env.Assignment(":=", 10)
	env.Assignment("+=", 10)
	env.Assignment("-=", 10)
	env.PostfixAssign("++", 10)
	env.PostfixAssign("--", 10)

	env.Infix("==", 40)
	env.Infix("!=", 40)
	env.Infix(">", 40)
	env.Infix(">=", 40)
	env.Infix("<", 40)
	env.Infix("<=", 40)

	arrayOp = &InfixOp{
		Bp:        80,
		MunchLeft: arrayOpMunchLeft,
	}
}

type RightMuncher func(env *Glisp, pr *Pratt) (Sexp, error)
type LeftMuncher func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error)

func InfixBuilder(env *Glisp, name string, args []Sexp) (Sexp, error) {
	//Q("InfixBuilder top, name='%s', len(args)==%v ", name, len(args))
	if name != "infixExpand" && len(args) != 1 {
		// let {} mean nil
		return SexpNull, nil
	}
	var arr *SexpArray
	//Q("InfixBuilder after top, args[0] has type ='%T' ", args[0])
	switch v := args[0].(type) {
	case *SexpArray:
		arr = v
	case *SexpPair:
		if name == "infixExpand" {
			pair, isPair := v.Tail.(*SexpPair)
			if !isPair {
				return SexpNull, fmt.Errorf("infixExpand expects (infix []) as its argument; instead we saw '%T'", v.Tail)
			}
			switch ar2 := pair.Head.(type) {
			case *SexpArray:
				//Q("infixExpand, doing recursive call to InfixBuilder, ar2 = '%v'", ar2.SexpString())
				return InfixBuilder(env, name, []Sexp{ar2})
			default:
				return SexpNull, fmt.Errorf("infixExpand expects (infix []) as its argument; instead we saw '%T'", v.Tail)
			}
		}
		return SexpNull, fmt.Errorf("InfixBuilder must receive an SexpArray")
	default:
		return SexpNull, fmt.Errorf("InfixBuilder must receive an SexpArray")
	}
	//Q("InfixBuilder, name='%s', arr = ", name)
	for i := range arr.Val {
		Q("arr[%v] = %v of type %T", i, arr.Val[i].SexpString(0), arr.Val[i])
	}
	pr := NewPratt(arr.Val)
	xs := []Sexp{}
	for {
		x, err := pr.Expression(env, 0)
		if err != nil {
			return SexpNull, err
		}
		if x == nil {
			Q("x was nil")
		} else {
			Q("x back is not nil and is of type %T/val = '%v', err = %v", x, x.SexpString(0), err)
		}
		if name == "infixExpand" {
			ret := MakeList([]Sexp{env.MakeSymbol("quote"), x})
			//Q("infixExpand: returning ret = '%v'", ret.SexpString(0))
			return ret, nil
		}
		_, isSemi := x.(*SexpSemicolon)
		if !isSemi {
			xs = append(xs, x)
		}
		Q("end of infix builder loop, pr.NextToken = '%v'", pr.NextToken.SexpString(0))
		if pr.IsEOF() {
			break
		}

	}
	//	Q("infix builder loop done, here are my expressions:")
	//	for i, ele := range xs {
	//		Q("xs[%v] = %v", i, ele.SexpString(0))
	//	}
	dup := env.Duplicate()
	ev, err := dup.EvalExpressions(xs)
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
// jea: Below is a working expression() parsing routine. Reproduces the
// original Pratt and Crockford formulation.
//
// AccumTree holds the accumulated parse tree at any point in time.
//     "The parse Tree Up to this point, by consuming the tokens
//      to the left" would be a better-but-too-long name.
//
//  and AccumTree is the stuff to the left of the
//   current operator in the parse stream.
//
// data flows from NextToken -> cnode -> (possibly on the stack of t
//   recursive MunchLeft calls) -> into the AccumTree tree.
//
//  better names: _left  -> AccumTree (to be returned)
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
		if ret == nil {
			Q("Expression is returning Sexp ret = nil")
		} else {
			Q("Expression is returning Sexp ret = '%v'", ret.SexpString(0))
		}
	}()
	cnode := p.NextToken
	if cnode != nil {
		Q("top of Expression, rbp = %v, cnode = '%v'", rbp, cnode.SexpString(0))
	} else {
		Q("top of Expression, rbp = %v, cnode is nil", rbp)
	}
	if p.IsEOF() {
		Q("Expression sees IsEOF, returning cnode = %v", cnode.SexpString(0))
		return cnode, nil
	}
	p.CnodeStack = append([]Sexp{p.NextToken}, p.CnodeStack...)
	//p.ShowCnodeStack()

	p.Advance()

	var curOp *InfixOp
	switch x := cnode.(type) {
	case *SexpSymbol:
		op, found := env.infixOps[x.name]
		if found {
			curOp = op
		}
	case *SexpArray:
		Q("in pratt parsing, got array x = '%v'", x.SexpString(0))
	}

	if curOp != nil && curOp.MunchRight != nil {
		// munch_right() of atoms returns this/itself, in which
		// case: p.AccumTree = t; is the result.
		Q("about to MunchRight on cnode = %v", cnode.SexpString(0))
		p.AccumTree, err = curOp.MunchRight(env, p)
		if err != nil {
			Q("Expression(%v) MunchRight saw err = %v", rbp, err)
			return SexpNull, err
		}
		Q("after MunchRight on cnode = %v, p.AccumTree = '%v'",
			cnode.SexpString(0), p.AccumTree.SexpString(0))
	} else {
		// do this, or have the default MunchRight return itself.
		p.AccumTree = cnode
	}

	for !p.IsEOF() {
		nextLbp := env.LeftBindingPower(p.NextToken)
		Q("nextLbp = %v", nextLbp)
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
		case *SexpArray:
			Q("assigning curOp to arrayOp")
			curOp = arrayOp
		default:
			panic(fmt.Errorf("how to handle cnode type = %#v", cnode))
		}

		p.CnodeStack[0] = p.NextToken
		//_cnode_stack.front() = NextToken;

		Q("in MunchLeft loop, before Advance, p.NextToken = %v",
			p.NextToken.SexpString(0))
		p.Advance()
		if p.Pos < len(p.Stream) {
			Q("in MunchLeft loop, after Advance, p.NextToken = %v",
				p.NextToken.SexpString(0))
		}

		// if cnode->munch_left() returns this/itself, then
		// the net effect is: p.AccumTree = cnode;
		if curOp != nil && curOp.MunchLeft != nil {
			Q("about to MunchLeft, cnode = %v, p.AccumTree = %v", cnode.SexpString(0), p.AccumTree.SexpString(0))
			p.AccumTree, err = curOp.MunchLeft(env, p, p.AccumTree)
			if err != nil {
				Q("curOp.MunchLeft saw err = %v", err)
				return SexpNull, err
			}
		} else {
			// do this, or have the default MunchLeft return itself.
			p.AccumTree = cnode
		}

	}

	p.CnodeStack = p.CnodeStack[1:]
	//_cnode_stack.pop_front()
	Q("at end of Expression(%v), returning p.AccumTree=%v, err=nil", rbp, p.AccumTree.SexpString(0))
	return p.AccumTree, nil
}

// Advance sets p.NextToken
func (p *Pratt) Advance() error {
	p.Pos++
	if p.Pos >= len(p.Stream) {
		return io.EOF
	}
	p.NextToken = p.Stream[p.Pos]
	Q("end of Advance, p.NextToken = '%v'", p.NextToken.SexpString(0))
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
		Q("LeftBindingPower: no entry in env.infixOps for operation '%s'",
			x.name)
		return 0
	case *SexpArray:
		return 80
	case *SexpSemicolon:
		return 0
	}
	panic(fmt.Errorf("LeftBindingPower: unhandled sx :%#v", sx))
	//return 0
}

func (p *Pratt) ShowCnodeStack() {
	if len(p.CnodeStack) == 0 {
		fmt.Println("CnodeStack is: empty")
		return
	}
	fmt.Println("CnodeStack is:")
	for i := range p.CnodeStack {
		fmt.Printf("CnodeStack[%v] = %v\n", i, p.CnodeStack[i].SexpString(0))
	}
}
