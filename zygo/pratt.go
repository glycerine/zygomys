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
func (env *Zlisp) Infix(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
			right, err := pr.Expression(env, bp)
			if err != nil {
				return SexpNull, err
			}
			list := MakeList([]Sexp{
				oper, left, right,
			})

			//Q("in Infix(), MunchLeft() call, pr.NextToken = %v. list returned = '%v'",
			//	pr.NextToken.SexpString(nil), list.SexpString(nil))
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

func (env *Zlisp) InfixF(op string, bp int, f func(env *Zlisp, op string, bp int) *InfixOp) *InfixOp {
	return f(env, op, bp)
}

// Infix creates a new (right-associative) short-circuiting
// infix operator, used for `and` and `or` in infix processing.
func (env *Zlisp) Infixr(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
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
func (env *Zlisp) Prefix(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchRight: func(env *Zlisp, pr *Pratt) (Sexp, error) {
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
func (env *Zlisp) Assignment(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	operSet := env.MakeSymbol("set")
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
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
			//Q("assignment returning list: '%v'", list.SexpString(nil))
			return list, nil
		},
		IsAssign: true,
	}
	env.infixOps[op] = iop
	return iop
}

// PostfixAssign creates a new postfix assignment operator for infix
// processing.
func (env *Zlisp) PostfixAssign(op string, bp int) *InfixOp {
	oper := env.MakeSymbol(op)
	iop := &InfixOp{
		Sym: oper,
		Bp:  bp,
		MunchLeft: func(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
			// TODO: check that left is okay as an LVALUE
			list := MakeList([]Sexp{
				oper, left,
			})
			//Q("postfix assignment returning list: '%v'", list.SexpString(nil))
			return list, nil
		},
	}
	env.infixOps[op] = iop
	return iop
}

func arrayOpMunchLeft(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
	oper := env.MakeSymbol("arrayidx")
	//Q("pr.NextToken = '%v', left = %#v", pr.NextToken.SexpString(nil), left)
	//if len(pr.CnodeStack) > 0 {
	//	Q("pr.CnodeStack[0] = '%v'", pr.CnodeStack[0])
	//}

	//right := pr.NextToken
	//Q("right = %#v", right)
	list := MakeList([]Sexp{
		oper, left, pr.CnodeStack[0],
	})
	return list, nil
}

func dotOpMunchLeft(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error) {
	//Q("dotOp MunchLeft, left = '%v'. NextToken='%v'. pr.CnodeStack[0]='%v'", left.SexpString(nil), pr.NextToken.SexpString(nil), pr.CnodeStack[0].SexpString(nil))
	list := MakeList([]Sexp{
		env.MakeSymbol("hashidx"), left, pr.CnodeStack[0],
	})
	return list, nil
}

func starOpMunchRight(env *Zlisp, pr *Pratt) (Sexp, error) {
	right, err := pr.Expression(env, 70)
	if err != nil {
		return SexpNull, err
	}
	list := MakeList([]Sexp{
		env.MakeSymbol("*"), right,
	})
	return list, nil
}

func isSymbolNamed(sx Sexp, name string) bool {
	sym, ok := sx.(*SexpSymbol)
	return ok && sym.name == name
}

func isInfixBlock(sx Sexp) bool {
	pair, ok := sx.(*SexpPair)
	if !ok {
		return false
	}
	return isSymbolNamed(pair.Head, "infix")
}

func isEmptyHashBlock(sx Sexp) bool {
	hash, ok := sx.(*SexpHash)
	return ok && hash.NumKeys == 0
}

func isForBodyBlock(sx Sexp) bool {
	return isInfixBlock(sx) || isEmptyHashBlock(sx)
}

func infixBlockIsEmpty(block Sexp) bool {
	_, empty, err := InfixArgsToArray("infixExpand", []Sexp{block})
	return err == nil && empty
}

func forBodyExpressions(block Sexp) ([]Sexp, error) {
	if isEmptyHashBlock(block) {
		return nil, nil
	}
	if !isInfixBlock(block) {
		return nil, fmt.Errorf("go-style for: missing body block")
	}
	if infixBlockIsEmpty(block) {
		return nil, nil
	}
	return []Sexp{block}, nil
}

func prattCall(env *Zlisp, name string, args ...Sexp) Sexp {
	return MakeList(append([]Sexp{env.MakeSymbol(name)}, args...))
}

func prattForControl(env *Zlisp, init, test, post Sexp) *SexpArray {
	return &SexpArray{Val: []Sexp{init, test, post}, Env: env}
}

func prattForList(env *Zlisp, label *SexpSymbol, control *SexpArray, body []Sexp) Sexp {
	args := []Sexp{env.MakeSymbol("for")}
	if label != nil {
		args = append(args, label)
	}
	args = append(args, control)
	args = append(args, body...)
	return MakeList(args)
}

func parsePrattOne(env *Zlisp, tokens []Sexp, context string) (Sexp, error) {
	if len(tokens) == 0 {
		return SexpNull, nil
	}
	pr := NewPratt(tokens)
	expr, err := pr.Expression(env, 0)
	if err != nil {
		return SexpNull, err
	}
	if !pr.IsEOF() {
		return SexpNull, fmt.Errorf("%s must be a single expression", context)
	}
	return expr, nil
}

func splitOnSemicolons(tokens []Sexp) [][]Sexp {
	segments := [][]Sexp{{}}
	for _, tok := range tokens {
		if _, isSemi := tok.(*SexpSemicolon); isSemi {
			segments = append(segments, []Sexp{})
			continue
		}
		last := len(segments) - 1
		segments[last] = append(segments[last], tok)
	}
	return segments
}

func countSemicolons(tokens []Sexp) int {
	n := 0
	for _, tok := range tokens {
		if _, isSemi := tok.(*SexpSemicolon); isSemi {
			n++
		}
	}
	return n
}

func hasRangeSymbol(tokens []Sexp) bool {
	for _, tok := range tokens {
		if isSymbolNamed(tok, "range") {
			return true
		}
	}
	return false
}

func findRangeAssign(tokens []Sexp) (int, string, bool) {
	for i, tok := range tokens {
		if isSymbolNamed(tok, ":=") {
			return i, ":=", true
		}
		if isSymbolNamed(tok, "=") {
			return i, "=", true
		}
	}
	return -1, "", false
}

func parseRangeTargets(tokens []Sexp) ([]*SexpSymbol, error) {
	if len(tokens) == 1 {
		sym, ok := tokens[0].(*SexpSymbol)
		if !ok {
			return nil, fmt.Errorf("go-style for range header: range target must be a symbol")
		}
		return []*SexpSymbol{sym}, nil
	}
	if len(tokens) == 3 {
		left, ok := tokens[0].(*SexpSymbol)
		if !ok {
			return nil, fmt.Errorf("go-style for range header: first range target must be a symbol")
		}
		if _, ok := tokens[1].(*SexpComma); !ok {
			return nil, fmt.Errorf("go-style for range header: two range targets must be separated by comma")
		}
		right, ok := tokens[2].(*SexpSymbol)
		if !ok {
			return nil, fmt.Errorf("go-style for range header: second range target must be a symbol")
		}
		return []*SexpSymbol{left, right}, nil
	}
	return nil, fmt.Errorf("go-style for range header: expected one or two range targets")
}

func lowerRangeBinding(env *Zlisp, targets []*SexpSymbol, define bool, sourceSym, indexSym *SexpSymbol, body []Sexp) Sexp {
	if len(targets) == 1 {
		op := "set"
		if define {
			op = "def"
		}
		return prattCall(env, op, targets[0], prattCall(env, "__rangeKey", sourceSym, indexSym))
	}

	pair := prattCall(env, "__rangePair", sourceSym, indexSym)
	if define {
		return prattCall(env, "mdef", targets[0], targets[1], pair)
	}

	pairSym := env.GenSymbol("__range_pair")
	pairBindings := &SexpArray{Val: []Sexp{pairSym, pair}, Env: env}
	beginArgs := []Sexp{
		env.MakeSymbol("begin"),
		prattCall(env, "set", targets[0], prattCall(env, "first", pairSym)),
		prattCall(env, "set", targets[1], prattCall(env, "second", pairSym)),
	}
	beginArgs = append(beginArgs, body...)
	return prattCall(env, "let", pairBindings, MakeList(beginArgs))
}

func lowerRangeFor(env *Zlisp, label *SexpSymbol, header []Sexp, body []Sexp) (Sexp, bool, error) {
	assignPos, op, foundAssign := findRangeAssign(header)
	if !foundAssign {
		if hasRangeSymbol(header) {
			return SexpNull, true, fmt.Errorf("go-style for range header: expected := or = before range")
		}
		return SexpNull, false, nil
	}

	if len(header) <= assignPos+1 || !isSymbolNamed(header[assignPos+1], "range") {
		if hasRangeSymbol(header) {
			return SexpNull, true, fmt.Errorf("go-style for range header: malformed range header")
		}
		return SexpNull, false, nil
	}

	targets, err := parseRangeTargets(header[:assignPos])
	if err != nil {
		return SexpNull, true, err
	}
	sourceTokens := header[assignPos+2:]
	if len(sourceTokens) == 0 {
		return SexpNull, true, fmt.Errorf("go-style for range header: missing range expression")
	}
	source, err := parsePrattOne(env, sourceTokens, "go-style for range expression")
	if err != nil {
		return SexpNull, true, err
	}

	sourceSym := env.GenSymbol("__range_src")
	lenSym := env.GenSymbol("__range_len")
	indexSym := env.GenSymbol("__range_i")
	letBindings := &SexpArray{
		Val: []Sexp{
			sourceSym, source,
			lenSym, prattCall(env, "__rangeLen", sourceSym),
		},
		Env: env,
	}
	control := prattForControl(env,
		prattCall(env, "def", indexSym, &SexpInt{Val: 0}),
		prattCall(env, "<", indexSym, lenSym),
		prattCall(env, "set", indexSym, prattCall(env, "+", indexSym, &SexpInt{Val: 1})),
	)

	binding := lowerRangeBinding(env, targets, op == ":=", sourceSym, indexSym, body)
	forBody := []Sexp{binding}
	if !(len(targets) == 2 && op == "=") {
		forBody = append(forBody, body...)
	}
	forLoop := prattForList(env, label, control, forBody)
	return prattCall(env, "letseq", letBindings, forLoop), true, nil
}

func lowerGoFor(env *Zlisp, label *SexpSymbol, header []Sexp, bodyBlock Sexp) (Sexp, error) {
	body, err := forBodyExpressions(bodyBlock)
	if err != nil {
		return SexpNull, err
	}

	nsemi := countSemicolons(header)
	if nsemi > 0 {
		if nsemi != 2 {
			return SexpNull, fmt.Errorf("go-style three-clause for header must contain exactly two semicolons")
		}
		segments := splitOnSemicolons(header)
		if len(segments) != 3 {
			return SexpNull, fmt.Errorf("go-style three-clause for header must contain init, condition, and post clauses")
		}
		init, err := parsePrattOne(env, segments[0], "go-style for init clause")
		if err != nil {
			return SexpNull, err
		}
		test := Sexp(&SexpBool{Val: true})
		if len(segments[1]) > 0 {
			test, err = parsePrattOne(env, segments[1], "go-style for condition clause")
			if err != nil {
				return SexpNull, err
			}
		}
		post, err := parsePrattOne(env, segments[2], "go-style for post clause")
		if err != nil {
			return SexpNull, err
		}
		control := prattForControl(env, init, test, post)
		return prattForList(env, label, control, body), nil
	}

	rangeLoop, isRange, err := lowerRangeFor(env, label, header, body)
	if err != nil || isRange {
		return rangeLoop, err
	}

	init := SexpNull
	test := Sexp(&SexpBool{Val: true})
	post := SexpNull
	if len(header) > 0 {
		test, err = parsePrattOne(env, header, "go-style for condition")
		if err != nil {
			return SexpNull, err
		}
	}
	control := prattForControl(env, init, test, post)
	return prattForList(env, label, control, body), nil
}

func forOpMunchRightWithLabel(env *Zlisp, pr *Pratt, label *SexpSymbol) (Sexp, error) {
	bodyPos := -1
	for i := pr.Pos; i < len(pr.Stream); i++ {
		if isForBodyBlock(pr.Stream[i]) {
			bodyPos = i
			break
		}
	}
	if bodyPos < 0 {
		return SexpNull, fmt.Errorf("go-style for: missing body block")
	}

	header := pr.Stream[pr.Pos:bodyPos]
	body := pr.Stream[bodyPos]
	pr.Pos = bodyPos
	_ = pr.Advance()
	return lowerGoFor(env, label, header, body)
}

func forOpMunchRight(env *Zlisp, pr *Pratt) (Sexp, error) {
	return forOpMunchRightWithLabel(env, pr, nil)
}

func (p *Pratt) LabeledFor(env *Zlisp) (Sexp, bool, error) {
	label, ok := p.NextToken.(*SexpSymbol)
	if !ok || !label.colonTail {
		return SexpNull, false, nil
	}
	if p.Pos+1 >= len(p.Stream) || !isSymbolNamed(p.Stream[p.Pos+1], "for") {
		return SexpNull, false, nil
	}

	_ = p.Advance()
	_ = p.Advance()
	if p.IsEOF() {
		return SexpNull, true, fmt.Errorf("go-style for: missing body block")
	}
	x, err := forOpMunchRightWithLabel(env, p, label)
	return x, true, err
}

func loopControlOpMunchRight(name string) RightMuncher {
	return func(env *Zlisp, pr *Pratt) (Sexp, error) {
		args := []Sexp{env.MakeSymbol(name)}
		if !pr.IsEOF() {
			if label, ok := pr.NextToken.(*SexpSymbol); ok {
				args = append(args, label)
				_ = pr.Advance()
			}
		}
		return MakeList(args), nil
	}
}

var arrayOp *InfixOp

// InitInfixOps establishes the env.infixOps definitions
// required for infix parsing using the Pratt parser.
func (env *Zlisp) InitInfixOps() {
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
	breakOp := env.Prefix("break", 0)
	breakOp.MunchRight = loopControlOpMunchRight("break")
	continueOp := env.Prefix("continue", 0)
	continueOp.MunchRight = loopControlOpMunchRight("continue")
	forOp := env.Prefix("for", 0)
	forOp.MunchRight = forOpMunchRight
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
	//Q("ifOp = %#v", ifOp.SexpString(nil))

	ifOp.MunchRight = func(env *Zlisp, pr *Pratt) (Sexp, error) {
		//Q("ifOp.MunchRight(): NextToken='%v'. pr.CnodeStack[0]='%v'", pr.NextToken.SexpString(nil), pr.CnodeStack[0].SexpString(nil))
		right, err := pr.Expression(env, 5)
		//Q("ifOp.MunchRight: back from Expression-1st-call, err = %#v, right = '%v'", err, right.SexpString(nil))
		if err != nil {
			return SexpNull, err
		}
		//Q("in ifOpMunchRight, got from p.Expression(env, 0) the right = '%v', err = %#v, pr.CnodeStack[0] = %#v, ifOp.Sym = '%v'", right.SexpString(nil), err, pr.CnodeStack[0], ifOp.Sym.SexpString(nil))

		thenExpr, err := pr.Expression(env, 0)
		//Q("ifOp.MunchRight: back from Expression-2nd-call, err = %#v, thenExpr = '%v'", err, thenExpr.SexpString(nil))
		if err != nil {
			return SexpNull, err
		}

		//Q("ifOp.MunchRight(), after Expression-2nd-call: . NextToken='%v'. pr.CnodeStack[0]='%v'", pr.NextToken.SexpString(nil), pr.CnodeStack[0].SexpString(nil))
		var elseExpr Sexp = SexpNull
		switch sym := pr.NextToken.(type) {
		case *SexpSymbol:
			if sym.name == "else" {
				//Q("detected else, advancing past it")
				pr.Advance()
				elseExpr, err = pr.Expression(env, 0)
				//Q("ifOp.MunchRight: back from Expression-3rd-call, err = %#v, elseExpr = '%v'", err, elseExpr.SexpString(nil))
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

type RightMuncher func(env *Zlisp, pr *Pratt) (Sexp, error)
type LeftMuncher func(env *Zlisp, pr *Pratt, left Sexp) (Sexp, error)
type StmtMuncher func(env *Zlisp, pr *Pratt) (Sexp, error)

func InfixArgsToArray(name string, args []Sexp) (*SexpArray, bool, error) {
	if name != "infixExpand" && len(args) != 1 {
		// let {} mean nil
		return nil, true, nil
	}
	var arr *SexpArray
	//Q("InfixBuilder after top, args[0] has type ='%T' ", args[0])
	switch v := args[0].(type) {
	case *SexpArray:
		arr = v
	case *SexpPair:
		if name == "infixExpand" {
			_, isSent := v.Tail.(*SexpSentinel)
			if isSent {
				// expansion of {} is nil
				return nil, true, nil
			}
			pair, isPair := v.Tail.(*SexpPair)
			if !isPair {
				return nil, false, fmt.Errorf("infixExpand expects (infix []) as its argument; instead we saw '%T' [err 3]", v.Tail)
			}
			switch ar2 := pair.Head.(type) {
			case *SexpArray:
				//Q("infixExpand, doing recursive call to InfixBuilder, ar2 = '%v'", ar2.SexpString(nil))
				return ar2, false, nil
			default:
				return nil, false, fmt.Errorf("infixExpand expects (infix []) as its argument; instead we saw '%T'", v.Tail)
			}
		}
		return nil, false, fmt.Errorf("InfixBuilder must receive an SexpArray. Saw: name='%v' / args[0]='%#v'", name, args[0])
	case *SexpHash:
		// an empty basic block {} that turned into an empty hash.
		return nil, true, nil
	default:
		return nil, false, fmt.Errorf("InfixBuilder (default) must receive an SexpArray. Saw: name='%v' / args[0]='%#v'", name, args[0])
	}
	return arr, false, nil
}

func InfixExpandArray(env *Zlisp, arr *SexpArray) ([]Sexp, error) {
	//Q("InfixBuilder, name='%s', arr = ", name)
	//for i := range arr.Val {
	//	Q("arr[%v] = '%v', of type %T", i, arr.Val[i].SexpString(nil), arr.Val[i])
	//}
	pr := NewPratt(arr.Val)
	xs := []Sexp{}

	for {
		x, ok, err := pr.LabeledFor(env)
		if err != nil {
			return nil, err
		}
		if !ok {
			x, err = pr.Expression(env, 0)
		}
		if err != nil {
			return nil, err
		}
		//if x == nil {
		//	Q("x was nil")
		//} else {
		//	Q("x back is not nil and is of type %T/val = '%v', err = %v", x, x.SexpString(nil), err)
		//}
		_, isSemi := x.(*SexpSemicolon)
		if !isSemi {
			xs = append(xs, x)
		}
		//Q("end of infix builder loop, pr.NextToken = '%v'", pr.NextToken.SexpString(nil))
		if pr.IsEOF() {
			break
		}

		_, nextIsSemi := pr.NextToken.(*SexpSemicolon)
		if nextIsSemi {
			pr.Advance() // skip over the semicolon
		}
	}
	return xs, nil
}

func InfixBuilder(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	//Q("InfixBuilder top, name='%s', len(args)==%v ", name, len(args))
	arr, empty, err := InfixArgsToArray(name, args)
	if err != nil {
		return SexpNull, err
	}
	if empty {
		return SexpNull, nil
	}

	xs, err := InfixExpandArray(env, arr)
	if err != nil {
		return SexpNull, err
	}
	//Q("infix builder loop done, here are my expressions:")
	//for i, ele := range xs {
	//	Q("xs[%v] = %v", i, ele.SexpString(nil))
	//}

	if name == "infixExpand" {
		ret := MakeList(append([]Sexp{env.MakeSymbol("quote")}, xs...))
		//Q("infixExpand: returning ret = '%v'", ret.SexpString(nil))
		return ret, nil
	}

	ev, err := EvalFunction(env, "infixEval", xs)
	if err != nil {
		return SexpNull, err
	}
	return ev, nil
}

type Pratt struct {
	NextToken  Sexp
	CnodeStack []Sexp
	AccumTree  Sexp

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

func (p *Pratt) Expression(env *Zlisp, rbp int) (ret Sexp, err error) {
	//defer func() {
	//	if ret == nil {
	//		Q("Expression is returning Sexp ret = nil")
	//	} else {
	//		Q("Expression is returning Sexp ret = '%v'", ret.SexpString(nil))
	//	}
	//}()

	cnode := p.NextToken
	//if cnode != nil {
	//	Q("top of Expression, rbp = %v, cnode = '%v'", rbp, cnode.SexpString(nil))
	//} else {
	//	Q("top of Expression, rbp = %v, cnode is nil", rbp)
	//}
	if p.IsEOF() {
		//Q("Expression sees IsEOF, returning cnode = %v", cnode.SexpString(nil))
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
			//    Q("Expression lookup of op.Sym=%v/op='%#v' succeeded", op.Sym.SexpString(nil), op)
			curOp = op
			//} else {
			//	Q("Expression lookup of x.name == '%v' failed", x.name)
		}
	case *SexpArray:
		//Q("in pratt parsing, got array x = '%v'", x.SexpString(nil))
	}

	if curOp != nil && curOp.MunchRight != nil {
		// munch_right() of atoms returns this/itself, in which
		// case: p.AccumTree = t; is the result.
		//Q("about to MunchRight on cnode = %v", cnode.SexpString(nil))
		p.AccumTree, err = curOp.MunchRight(env, p)
		if err != nil {
			//Q("Expression(%v) MunchRight saw err = %v", rbp, err)
			return SexpNull, err
		}
		//Q("after MunchRight on cnode = %v, p.AccumTree = '%v'",
		//	cnode.SexpString(nil), p.AccumTree.SexpString(nil))
	} else {
		// do this, or have the default MunchRight return itself.
		p.AccumTree = cnode
	}

	for !p.IsEOF() {
		nextLbp, err := env.LeftBindingPower(p.NextToken)
		if err != nil {
			//Q("env.LeftBindingPower('%s') saw err = %v",
			//	p.NextToken.SexpString(nil), err)
			return SexpNull, err
		}
		//Q("nextLbp = %v, and rbp = %v, so rpb >= nextLbp == %v", nextLbp, rbp, rbp >= nextLbp)
		if rbp >= nextLbp {
			//Q("found rbp >= nextLbp so breaking out of left-binding loop")
			break
		}

		cnode = p.NextToken
		curOp = nil
		switch x := cnode.(type) {
		case *SexpSymbol:
			op, found := env.infixOps[x.name]
			if found {
				//Q("assigning curOp <- cnode '%s'", x.name)
				curOp = op
			} else {
				if x.isDot {
					curOp = env.infixOps["."]
					//Q("assigning curOp <- dotInfixOp; then curOp = %#v", curOp)
				}
			}
		case *SexpArray:
			//Q("assigning curOp <- arrayOp")
			curOp = arrayOp
		case *SexpComma:
			curOp = env.infixOps["comma"]
			//Q("assigning curOp <- infixOps[`comma`]; then curOp = %#v", curOp)
		case *SexpPair:
			// sexp-call, treat like function call with rbp 80
			//Q("Expression sees an SexpPair")
			// leaving curOp nil seems to work just fine here.
		default:
			panic(fmt.Errorf("how to handle cnode type = %#v", cnode))
		}
		//Q("curOp = %#v", curOp)

		p.CnodeStack[0] = p.NextToken
		//_cnode_stack.front() = NextToken;

		//Q("in MunchLeft loop, before Advance, p.NextToken = %v",
		//	p.NextToken.SexpString(nil))
		p.Advance()
		if p.Pos < len(p.Stream) {
			//Q("in MunchLeft loop, after Advance, p.NextToken = %v",
			//	p.NextToken.SexpString(nil))
		}

		// if cnode->munch_left() returns this/itself, then
		// the net effect is: p.AccumTree = cnode;
		if curOp != nil && curOp.MunchLeft != nil {
			//Q("about to MunchLeft, cnode = %v, p.AccumTree = %v", cnode.SexpString(nil), p.AccumTree.SexpString(nil))
			p.AccumTree, err = curOp.MunchLeft(env, p, p.AccumTree)
			if err != nil {
				//Q("curOp.MunchLeft saw err = %v", err)
				return SexpNull, err
			}
		} else {
			//Q("curOp has not MunchLeft, setting AccumTree <- cnode. here cnode = %v", cnode.SexpString(nil))
			// do this, or have the default MunchLeft return itself.
			p.AccumTree = cnode
		}

	} // end for !p.IsEOF()

	p.CnodeStack = p.CnodeStack[1:]
	//_cnode_stack.pop_front()
	//Q("at end of Expression(%v), returning p.AccumTree=%v, err=nil", rbp, p.AccumTree.SexpString(nil))
	return p.AccumTree, nil
}

// Advance sets p.NextToken
func (p *Pratt) Advance() error {
	p.Pos++
	if p.Pos >= len(p.Stream) {
		return io.EOF
	}
	p.NextToken = p.Stream[p.Pos]
	//Q("end of Advance, p.NextToken = '%v'", p.NextToken.SexpString(nil))
	return nil
}

func (p *Pratt) IsEOF() bool {
	if p.Pos >= len(p.Stream) {
		return true
	}
	return false
}

func (env *Zlisp) LeftBindingPower(sx Sexp) (int, error) {
	//Q("LeftBindingPower: sx is '%v'", sx.SexpString(nil))
	switch x := sx.(type) {
	case *SexpInt, *SexpFloat:
		return 0, nil
	case *SexpBool:
		return 0, nil
	case *SexpStr:
		return 0, nil
	case *SexpSymbol:
		op, found := env.infixOps[x.name]
		if x.name == "if" {
			// we don't want if to be doing any binding to the left,
			// so we enforce that it has zero left-binding power. It
			// gets a right-binding power of 5 since it is a prefix operator.
			//Q("LeftBindingPower: found if, return 0 left-binding-power")
			return 0, nil
		}
		if found {
			//Q("LeftBindingPower: found op '%#v', returning op.Bp = %v", op, op.Bp)
			return op.Bp, nil
		}
		if x.isDot {
			//Q("LeftBindingPower: dot symbol '%v', "+
			//	"giving it binding-power 80", x.name)
			return 80, nil
		}
		//Q("LeftBindingPower: no entry in env.infixOps for operation '%s'", x.name)
		return 0, nil
	case *SexpArray:
		return 80, nil
	case *SexpComma:
		return 15, nil
	case *SexpSemicolon:
		return 0, nil
	case *SexpComment:
		return 0, nil
	case *SexpPair:
		if x.Head != nil {
			switch sym := x.Head.(type) {
			case *SexpSymbol:
				if sym.name == "infix" {
					//Q("detected infix!!! -- setting binding power to 0")
					return 0, nil
				}
			}
		}
		return 0, nil
	case *SexpHash:
		// an empty {} block that became an empty hash. no-op.
		return 0, nil
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
		fmt.Printf("CnodeStack[%v] = %v\n", i, p.CnodeStack[i].SexpString(nil))
	}
}
