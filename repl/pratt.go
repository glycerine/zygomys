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
	MunchStmt  StmtMuncher  // aka std. Used only at the beginning of a statement.
	IsAssign   bool
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

			Q("in Infix(), MunchLeft() call, pr.NextToken = %v. list returned = '%v'",
				pr.NextToken.SexpString(0), list.SexpString(0))
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
			// TODO: check that left is okay as an LVALUE.

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
		IsAssign: true,
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

func dotOpMunchLeft(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
	//Q("dotOp MunchLeft, left = '%v'. NextToken='%v'. pr.CnodeStack[0]='%v'", left.SexpString(0), pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))
	list := MakeList([]Sexp{
		env.MakeSymbol("hashidx"), left, pr.CnodeStack[0],
	})
	return list, nil
}

func starOpMunchRight(env *Glisp, pr *Pratt) (Sexp, error) {
	right, err := pr.Expression(env, 70)
	if err != nil {
		return SexpNull, err
	}
	list := MakeList([]Sexp{
		env.MakeSymbol("*"), right,
	})
	return list, nil
}

var arrayOp *InfixOp

// InitInfixOps establishes the env.infixOps definitions
// required for infix parsing using the Pratt parser.
func (env *Glisp) InitInfixOps() {
	env.Infix("+", 50)
	env.Infix("-", 50)

	star := env.Infix("*", 60)
	star.MunchRight = starOpMunchRight

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

	// set the global arrayOp
	arrayOp = &InfixOp{
		Bp:        80,
		MunchLeft: arrayOpMunchLeft,
	}

	dotOp := env.Infix(".", 80)
	dotOp.MunchLeft = dotOpMunchLeft

	ifOp := env.Prefix("if", 5)
	Q("ipOf = %#v", ifOp)

	ifOp.MunchRight = func(env *Glisp, pr *Pratt) (Sexp, error) {
		Q("ifOp.MunchRight(): NextToken='%v'. pr.CnodeStack[0]='%v'", pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))
		right, err := pr.Expression(env, 0)
		Q("ifOp.MunchRight: back from Expression-1st-call, err = %#v, right = '%v'", err, right.SexpString(0))
		if err != nil {
			return SexpNull, err
		}
		Q("in ifOpMunchRight, got from p.Expression(env, 0) the right = '%v', err = %#v, pr.CnodeStack[0] = %#v, ifOp.Sym = '%v'", right.SexpString(0), err, pr.CnodeStack[0], ifOp.Sym.SexpString(0))

		thenExpr, err := pr.Expression(env, 0)
		Q("ifOp.MunchRight: back from Expression-2nd-call, err = %#v, thenExpr = '%v'", err, thenExpr.SexpString(0))
		if err != nil {
			return SexpNull, err
		}

		Q("ifOp.MunchRight(), after Expression-2nd-call: . NextToken='%v'. pr.CnodeStack[0]='%v'", pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))
		var elseExpr Sexp = SexpNull
		switch sym := pr.NextToken.(type) {
		case *SexpSymbol:
			if sym.name == "else" {
				Q("detected else, advancing past it")
				pr.Advance()
				elseExpr, err = pr.Expression(env, 0)
				Q("ifOp.MunchRight: back from Expression-3rd-call, err = %#v, elseExpr = '%v'", err, elseExpr.SexpString(0))
				if err != nil {
					return SexpNull, err
				}
			}
		}

		list := MakeList([]Sexp{
			env.MakeSymbol("cond"), right, thenExpr, elseExpr,
		})
		return list, nil
	}

	env.Infix("comma", 15)
}

func ifOpMunchStmt(env *Glisp, pr *Pratt) (Sexp, error) {
	Q("in ifOpMunchStmt() NextToken='%v'. pr.CnodeStack[0]='%v'",
		pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))
	list := MakeList([]Sexp{
		env.MakeSymbol("ifOpMunchStmtOutput"), pr.CnodeStack[0],
	})
	Q("ifOpMunchStmt() returning list = '%#v'", list)
	return list, nil
}

func ifOpMunchLeft(env *Glisp, pr *Pratt, left Sexp) (Sexp, error) {
	Q("ifOpMunchLeft(), left = '%v'. NextToken='%v'. pr.CnodeStack[0]='%v'", left.SexpString(0), pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))
	list := MakeList([]Sexp{
		env.MakeSymbol("ifOpMunchLeftOutput"), left, pr.CnodeStack[0],
	})
	return list, nil
}

func ifOpMunchRight(env *Glisp, pr *Pratt) (Sexp, error) {
	Q("ifOpMunchRight(). NextToken='%v'. pr.CnodeStack[0]='%v'", pr.NextToken.SexpString(0), pr.CnodeStack[0].SexpString(0))

	right, err := pr.Expression(env, 0)
	if err != nil {
		return SexpNull, err
	}
	Q("in ifOpMunchRight, got from p.Expression(env, 0) the right = %#v, err = %#v, pr.CnodeStack[0] = %#v",
		right, err, pr.CnodeStack[0])

	list := MakeList([]Sexp{
		//		env.MakeSymbol("ifOpMunchRightOutput"), right, pr.CnodeStack[0],
		env.MakeSymbol("ifOpMunchRightOutput"), right,
	})
	return list, nil
}

type RightMuncher func(env *Glisp, pr *Pratt) (Sexp, error)
type LeftMuncher func(env *Glisp, pr *Pratt, left Sexp) (Sexp, error)
type StmtMuncher func(env *Glisp, pr *Pratt) (Sexp, error)

func InfixBuilder(env *Glisp, name string, args []Sexp) (Sexp, error) {
	Q("InfixBuilder top, name='%s', len(args)==%v ", name, len(args))
	if name != "infixExpand" && len(args) != 1 {
		// let {} mean nil
		return SexpNull, nil
	}
	var arr *SexpArray
	Q("InfixBuilder after top, args[0] has type ='%T' ", args[0])
	switch v := args[0].(type) {
	case *SexpArray:
		arr = v
	case *SexpPair:
		if name == "infixExpand" {
			_, isSent := v.Tail.(*SexpSentinel)
			if isSent {
				// expansion of {} is nil
				return SexpNull, nil
			}
			pair, isPair := v.Tail.(*SexpPair)
			if !isPair {
				return SexpNull, fmt.Errorf("infixExpand expects (infix []) as its argument; instead we saw '%T' [err 3]", v.Tail)
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
	Q("InfixBuilder, name='%s', arr = ", name)
	for i := range arr.Val {
		Q("arr[%v] = '%v', of type %T", i, arr.Val[i].SexpString(0), arr.Val[i])
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
	Q("infix builder loop done, here are my expressions:")
	for i, ele := range xs {
		Q("xs[%v] = %v", i, ele.SexpString(0))
	}
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
// guarantee the precedence-hill-climbing property (small precedence
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
			Q("Expression lookup of op.Sym=%v/op='%#v' succeeded", op.Sym.SexpString(0), op)
			curOp = op
		} else {
			Q("Expression lookup of x.name == '%v' failed", x.name)
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
		Q("not IsEOF in Expression")
		nextLbp, err := env.LeftBindingPower(p.NextToken)
		if err != nil {
			Q("env.LeftBindingPower('%s') saw err = %v",
				p.NextToken.SexpString(0), err)
			return SexpNull, err
		}
		Q("nextLbp = %v, and rbp = %v, so rpb >= nextLbp == %v", nextLbp, rbp, rbp >= nextLbp)
		if rbp >= nextLbp {
			break
		}

		cnode = p.NextToken
		curOp = nil
		switch x := cnode.(type) {
		case *SexpSymbol:
			op, found := env.infixOps[x.name]
			if found {
				Q("assigning curOp <- cnode '%s'", x.name)
				curOp = op
			} else {
				if x.isDot {
					curOp = env.infixOps["."]
					Q("assigning curOp <- dotInfixOp; then curOp = %#v", curOp)
				}
			}
		case *SexpArray:
			Q("assigning curOp <- arrayOp")
			curOp = arrayOp
		case *SexpComma:
			curOp = env.infixOps["comma"]
			Q("assigning curOp <- infixOps[`comma`]; then curOp = %#v", curOp)
		case *SexpPair:
			// sexp-call, treat like function call with rbp 80
			Q("Expression sees an SexpPair")
			// leaving curOp nil seems to work just fine here.
		default:
			panic(fmt.Errorf("how to handle cnode type = %#v", cnode))
		}
		Q("curOp = %#v", curOp)

		p.CnodeStack[0] = p.NextToken
		//_cnode_stack.front() = NextToken;

		Q("in MunchLeft loop, before Advance, p.NextToken = %v",
			p.NextToken.SexpString(0))
		p.Advance()
		if p.Pos < len(p.Stream) {
			Q("in MunchLeft loop, after Advance, p.NextToken = %v",
				p.NextToken.SexpString(0))
		}

		// dotOpMunchLeft?

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
			Q("curOp has not MunchLeft, setting AccumTree <- cnode")
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

func (env *Glisp) LeftBindingPower(sx Sexp) (int, error) {
	Q("LeftBindingPower: sx is '%#v'", sx)
	switch x := sx.(type) {
	case *SexpInt:
		return 0, nil
	case *SexpBool:
		return 0, nil
	case *SexpStr:
		return 0, nil
	case *SexpSymbol:
		op, found := env.infixOps[x.name]
		if found {
			return op.Bp, nil
		}
		if x.isDot {
			Q("LeftBindingPower: dot symbol '%v', "+
				"giving it binding-power 80", x.name)
			return 80, nil
		}
		Q("LeftBindingPower: no entry in env.infixOps for operation '%s'",
			x.name)
		return 0, nil
	case *SexpArray:
		return 80, nil
	case *SexpComma:
		return 15, nil
	case *SexpSemicolon:
		return 0, nil
	case *SexpPair:
		if x.Head != nil {
			switch sym := x.Head.(type) {
			case *SexpSymbol:
				if sym.name == "infix" {
					Q("detected infix!!! -- setting binding power to 0")
					return 0, nil
				}
			}
		}
		return 80, nil
	}
	return 0, fmt.Errorf("LeftBindingPower: unhandled sx :%#v", sx)
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

func (p *Pratt) Statement(env *Glisp) (ret Sexp, err error) {

	cnode := p.NextToken
	if cnode != nil {
		Q("top of Statement, cnode = '%v'", cnode.SexpString(0))
	} else {
		Q("top of Statement, cnode is nil")
	}
	if p.IsEOF() {
		Q("Statement sees IsEOF, returning cnode = %v", cnode.SexpString(0))
		return cnode, nil
	}

	var curOp *InfixOp
	switch x := cnode.(type) {
	case *SexpSymbol:
		Q("in pratt Statement() parsing, cnode = '%v'", cnode.SexpString(0))
		op, found := env.infixOps[x.name]
		if found {
			curOp = op
		} else {
			panic("unknown statement operation")
		}
	case *SexpArray:
		Q("in pratt Statement() parsing, got array x = '%v'", x.SexpString(0))
		panic("and do what?")
	}

	if curOp != nil && curOp.MunchStmt != nil {
		Q("curOp '%#v' has a MunchStmt", curOp)
		p.Advance()
		//scope.reserve(curOp)
		return curOp.MunchStmt(env, p)
	}
	v, err := p.Expression(env, 0)
	Q("in Statement, got from p.Expression(env, 0) the v, err = %#v, %#v", v, err)
	//	if (!v.IsAssignment && v.id !== "(") {
	//		v.error("Bad expression statement.");
	//	}
	//p.Advance(";");
	return v, err
}

/*
// suspect parser.go ParseInfix takes care of this.
func (p *Pratt) MultipleStatements(env *Glisp) (ret Sexp, err error) {
	var a []Sexp
	var s Sexp
	for !p.IsEOF() {

		Q("MultipleStatements: p.NextToken = %#v", p.NextToken)
		if p.NextToken.Sym != nil {
			//switch p.NextNoken.Sym.name == "}" {
			break
		}
		s, err = p.Statement(env)
		Q("MultipleStatements: back from p.Statement(env): s = '%v', err = '%v'", s.SexpString(0), err)
		if err != nil {
			break
		}
		if s != nil {
			a = append(a, s)
		} else {
			break
		}
	}
	return &SexpArray{Val: a}, err
}
*/

/*
var block = function () {
	var t = token;
	advance("{");
	return t.std();
};
*/

// Statement parses one statement (e.g. if-statement)

// Stmt creates a new statement
func (env *Glisp) Stmt(op string) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  80,
		MunchStmt: func(env *Glisp, pr *Pratt) (Sexp, error) {
			right, err := pr.Expression(env, 0)
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
