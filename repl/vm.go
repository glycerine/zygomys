package zygo

import (
	"errors"
	"fmt"
)

type Instruction interface {
	InstrString() string
	Execute(env *Glisp) error
}

type JumpInstr struct {
	addpc int
	where string
}

var OutOfBounds error = errors.New("jump out of bounds")

func (j JumpInstr) InstrString() string {
	return fmt.Sprintf("jump %d %s", j.addpc, j.where)
}

func (j JumpInstr) Execute(env *Glisp) error {
	newpc := env.pc + j.addpc
	if newpc < 0 || newpc > env.CurrentFunctionSize() {
		return OutOfBounds
	}
	env.pc = newpc
	return nil
}

type GotoInstr struct {
	location int
}

func (g GotoInstr) InstrString() string {
	return fmt.Sprintf("goto %d", g.location)
}

func (g GotoInstr) Execute(env *Glisp) error {
	if g.location < 0 || g.location > env.CurrentFunctionSize() {
		return OutOfBounds
	}
	env.pc = g.location
	return nil
}

type BranchInstr struct {
	direction bool
	location  int
}

func (b BranchInstr) InstrString() string {
	var format string
	if b.direction {
		format = "br %d"
	} else {
		format = "brn %d"
	}
	return fmt.Sprintf(format, b.location)
}

func (b BranchInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	if b.direction == IsTruthy(expr) {
		return JumpInstr{addpc: b.location}.Execute(env)
	}
	env.pc++
	return nil
}

type PushInstr struct {
	expr Sexp
}

func (p PushInstr) InstrString() string {
	return "push " + p.expr.SexpString(0)
}

func (p PushInstr) Execute(env *Glisp) error {
	env.datastack.PushExpr(p.expr)
	env.pc++
	return nil
}

type PopInstr int

func (p PopInstr) InstrString() string {
	return "pop"
}

func (p PopInstr) Execute(env *Glisp) error {
	_, err := env.datastack.PopExpr()
	env.pc++
	return err
}

type DupInstr int

func (d DupInstr) InstrString() string {
	return "dup"
}

func (d DupInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.GetExpr(0)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type EnvToStackInstr struct {
	sym *SexpSymbol
}

func (g EnvToStackInstr) InstrString() string {
	return fmt.Sprintf("envToStack %s", g.sym.name)
}

func (g EnvToStackInstr) Execute(env *Glisp) error {
	VPrintf("in EnvToStackInstr\n")
	defer VPrintf("leaving EnvToStackInstr env.pc =%v\n", env.pc)

	macxpr, isMacro := env.macros[g.sym.number]
	if isMacro {
		if macxpr.orig != nil {
			return fmt.Errorf("'%s' is a macro, with definition: %s\n", g.sym.name, macxpr.orig.SexpString(0))
		}
		return fmt.Errorf("'%s' is a builtin macro.\n", g.sym.name)
	}
	var expr Sexp
	var err error
	expr, err, _ = env.LexicalLookupSymbol(g.sym, false)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type PopStackPutEnvInstr struct {
	sym *SexpSymbol
}

func (p PopStackPutEnvInstr) InstrString() string {
	return fmt.Sprintf("popStackPutEnv %s", p.sym.name)
}

func (p PopStackPutEnvInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	env.pc++
	return env.LexicalBindSymbol(p.sym, expr)

}

// Update takes top of datastack and
// assigns it to sym when sym is found
// already in the current scope or
// up the stack. Used
// to implement (set v 10) when v is
// not in the local scope.
//
type UpdateInstr struct {
	sym *SexpSymbol
}

func (p UpdateInstr) InstrString() string {
	return fmt.Sprintf("putup %s", p.sym.name)
}

func (p UpdateInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	env.pc++
	var scope *Scope

	if p.sym.isSigil {
		Q("UpdateInstr: ignoring sigil symbol '%s'", p.sym.SexpString(0))
		return nil
	}
	if p.sym.isDot {
		Q("UpdateInstr: dot symbol '%s' being updated with dotGetSetHelper()",
			p.sym.SexpString(0))
		_, err := dotGetSetHelper(env, p.sym.name, &expr)
		return err
	}
	/*
		switch p.sym.(type) {
		case Selector:
			P("UpdateInstr: I see Selector '%s'", p.sym.SexpString(0))
		}
	*/
	_, err, scope = env.LexicalLookupSymbol(p.sym, false)
	if err != nil {
		// not found up the stack, so treat like (def)
		// instead of (set)
		return env.LexicalBindSymbol(p.sym, expr)
	}

	// found up the stack, so (set)
	return scope.UpdateSymbolInScope(p.sym, expr)
}

type CallInstr struct {
	sym   *SexpSymbol
	nargs int
}

func (c CallInstr) InstrString() string {
	return fmt.Sprintf("call %s %d", c.sym.name, c.nargs)
}

func (c CallInstr) Execute(env *Glisp) error {
	f, ok := env.builtins[c.sym.number]
	if ok {
		_, err := env.CallUserFunction(f, c.sym.name, c.nargs)
		return err
	}
	var funcobj, indirectFuncName Sexp
	var err error

	funcobj, err, _ = env.LexicalLookupSymbol(c.sym, false)

	if err != nil {
		return err
	}
	Q("\n in CallInstr, after looking up c.sym='%s', got funcobj='%v'. datastack is:\n", c.sym.name, funcobj.SexpString(0))
	//env.datastack.PrintStack()
	switch f := funcobj.(type) {
	case *SexpSymbol:
		// is it a dot-symbol call?
		Q("\n in CallInstr, found symbol\n")
		if c.sym.isDot {
			Q("\n in CallInstr, found symbol, c.sym.isDot is true\n")

			dotSymRef, dotLookupErr := dotGetSetHelper(env, c.sym.name, nil)
			// cannot error out yet, we might be assigning to a new field,
			// not already set.

			// are we a value request (no further args), or a fuction/method call?
			Q("\n in CallInstr, found dot-symbol\n")
			if c.nargs == 0 {
				// value request
				if dotLookupErr != nil {
					return dotLookupErr
				}
				env.datastack.PushExpr(dotSymRef)
				env.pc++
				return nil
			} else {
				// function call
				Q("in CallInstr, dot-symbol followed by more args")
				// get our function from the top of datastack
				expressions, err := env.datastack.PopExpressions(c.nargs)
				if err != nil {
					return err
				}

				// does our dot-symbol itself refer to a function?
				Q("in CallInstr, found dot-symbol, dot-symbol refers to type %T", dotSymRef)
				switch fn := dotSymRef.(type) {
				case *SexpFunction:
					c.setupDotCallHelper(env, fn, &indirectFuncName, expressions, 0, dotSymRef)
				default:
					top := expressions[0]
					Q("in CallInstr, found dot-symbol, first arg to dot-symbol is of type %T", top)
					switch ftop := top.(type) {
					case *SexpFunction:
						c.setupDotCallHelper(env, ftop, &indirectFuncName, expressions, 1, dotSymRef)
					default:
						return fmt.Errorf("dot-symbol '%s' was followed by non-function '%s'.",
							c.sym.name, ftop.SexpString(0))
					}
				}
			}
		} else {
			// not isDot

			// allow symbols to refer to dot-symbols, that then we call
			indirectFuncName, err = dotGetSetHelper(env, f.name, nil)
			if err != nil {
				return fmt.Errorf("'%s' refers to symbol '%s', but '%s' could not be resolved: '%s'.",
					c.sym.name, f.name, f.name, err)
			}

			// allow symbols to refer to functions that we then call
			/*
				indirectFuncName, err, _ = env.LexicalLookupSymbol(f, false)
				if err != nil {
					return fmt.Errorf("'%s' refers to symbol '%s', but '%s' could not be resolved: '%s'.",
						c.sym.name, f.name, f.name, err)
				}
			*/
			Q("\n in CallInstr, found symbol, c.sym.isDot is false. f of type %T/val = %v. indirectFuncName = '%v'\n", f, f.SexpString(0), indirectFuncName.SexpString(0))

		}

		switch g := indirectFuncName.(type) {
		case *SexpFunction:
			if !g.user {
				return env.CallFunction(g, c.nargs)
			}
			_, err := env.CallUserFunction(g, f.name, c.nargs)
			return err
		default:
			if err != nil {
				return fmt.Errorf("symbol '%s' refers to '%s' which does not refer to a function.", c.sym.name, f.name)
			}
		}

	case *SexpFunction:
		if !f.user {
			return env.CallFunction(f, c.nargs)
		}
		_, err := env.CallUserFunction(f, c.sym.name, c.nargs)
		return err

	case *RegisteredType:
		if f.Constructor == nil {
			env.pc++
			res, err := baseConstruct(env, f, c.nargs)
			if err != nil {
				return err
			}
			env.datastack.PushExpr(res)
			return nil
		}
		Q("call instruction for RegisteredType!")
		_, err := env.CallUserFunction(f.Constructor, c.sym.name, c.nargs)
		return err
	}
	return errors.New(fmt.Sprintf("%s is not a function", c.sym.name))
}

type DispatchInstr struct {
	nargs int
}

func (d DispatchInstr) InstrString() string {
	return fmt.Sprintf("dispatch %d", d.nargs)
}

func (d DispatchInstr) Execute(env *Glisp) error {
	funcobj, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	switch f := funcobj.(type) {
	case *SexpFunction:
		if !f.user {
			return env.CallFunction(f, d.nargs)
		}
		_, err := env.CallUserFunction(f, f.name, d.nargs)
		return err
	}
	// allow ([] int64) to express slice of int64.
	switch arr := funcobj.(type) {
	case *SexpArray:
		if len(arr.Val) == 0 {
			_, err := env.CallUserFunction(sxSliceOf, funcobj.SexpString(0), d.nargs)
			return err
		}
		// call along with the array as an argument so we know the size of the
		// array / matrix / tensor to make. The 2nd argument will be the dimension array.
		env.datastack.PushExpr(arr)
		_, err := env.CallUserFunction(sxArrayOf, funcobj.SexpString(0), d.nargs+1)
		return err
	}
	return fmt.Errorf("not a function on top of datastack: '%T/%#v'", funcobj, funcobj)
}

type ReturnInstr struct {
	err error
}

func (r ReturnInstr) Execute(env *Glisp) error {
	if r.err != nil {
		return r.err
	}
	return env.ReturnFromFunction()
}

func (r ReturnInstr) InstrString() string {
	if r.err == nil {
		return "ret"
	}
	return "ret \"" + r.err.Error() + "\""
}

type AddScopeInstr struct {
	Name string
}

func (a AddScopeInstr) InstrString() string {
	return "add scope " + a.Name
}

func (a AddScopeInstr) Execute(env *Glisp) error {
	sc := env.NewNamedScope(fmt.Sprintf("runtime add scope for '%s' at pc=%v",
		env.curfunc.name, env.pc))
	env.linearstack.Push(sc)
	env.pc++
	return nil
}

type AddFuncScopeInstr struct {
	Name string
}

func (a AddFuncScopeInstr) InstrString() string {
	return "add func scope " + a.Name
}

func (a AddFuncScopeInstr) Execute(env *Glisp) error {
	sc := env.NewNamedScope(fmt.Sprintf("%s at pc=%v",
		env.curfunc.name, env.pc))
	sc.IsFunction = true
	env.linearstack.Push(sc)
	env.pc++
	return nil
}

type RemoveScopeInstr struct{}

func (a RemoveScopeInstr) InstrString() string {
	return "rem runtime scope"
}

func (a RemoveScopeInstr) Execute(env *Glisp) error {
	env.pc++
	return env.linearstack.PopScope()
}

type ExplodeInstr int

func (e ExplodeInstr) InstrString() string {
	return "explode"
}

func (e ExplodeInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	arr, err := ListToArray(expr)
	if err != nil {
		return err
	}

	for _, val := range arr {
		env.datastack.PushExpr(val)
	}
	env.pc++
	return nil
}

type SquashInstr int

func (s SquashInstr) InstrString() string {
	return "squash"
}

func (s SquashInstr) Execute(env *Glisp) error {
	var list Sexp = SexpNull

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		list = Cons(expr, list)
	}
	env.datastack.PushExpr(list)
	env.pc++
	return nil
}

// bind these symbols to the SexpPair list found at
// datastack top.
type BindlistInstr struct {
	syms []*SexpSymbol
}

func (b BindlistInstr) InstrString() string {
	joined := ""
	for _, s := range b.syms {
		joined += s.name + " "
	}
	return fmt.Sprintf("bindlist %s", joined)
}

func (b BindlistInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}

	arr, err := ListToArray(expr)
	if err != nil {
		return err
	}

	nsym := len(b.syms)
	narr := len(arr)
	if narr < nsym {
		return fmt.Errorf("bindlist failing: %d targets but only %d sources", nsym, narr)
	}

	for i, bindThisSym := range b.syms {
		env.LexicalBindSymbol(bindThisSym, arr[i])
	}
	env.pc++
	return nil
}

type VectorizeInstr int

func (s VectorizeInstr) InstrString() string {
	return "vectorize"
}

func (s VectorizeInstr) Execute(env *Glisp) error {
	vec := make([]Sexp, 0)

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		vec = append([]Sexp{expr}, vec...)
	}
	env.datastack.PushExpr(&SexpArray{Val: vec, Env: env})
	env.pc++
	return nil
}

type HashizeInstr struct {
	HashLen  int
	TypeName string
}

func (s HashizeInstr) InstrString() string {
	return "hashize"
}

func (s HashizeInstr) Execute(env *Glisp) error {
	a := make([]Sexp, 0)

	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		if expr == SexpMarker {
			break
		}
		a = append(a, expr)
	}
	hash, err := MakeHash(a, s.TypeName, env)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(hash)
	env.pc++
	return nil
}

type LabelInstr struct {
	label string
}

func (s LabelInstr) InstrString() string {
	return fmt.Sprintf("label %s", s.label)
}

func (s LabelInstr) Execute(env *Glisp) error {
	env.pc++
	return nil
}

type BreakInstr struct {
	loop *Loop
	pos  int
}

func (s BreakInstr) InstrString() string {
	if s.pos == 0 {
		return fmt.Sprintf("break %s", s.loop.stmtname.name)
	}
	return fmt.Sprintf("break %s (loop is at %d)", s.loop.stmtname.name, s.pos)
}

func (s *BreakInstr) Execute(env *Glisp) error {
	if s.pos == 0 {
		pos, err := env.FindLoop(s.loop)
		if err != nil {
			return err
		}
		s.pos = pos
	}
	env.pc = s.pos + s.loop.breakOffset
	return nil
}

type ContinueInstr struct {
	loop *Loop
	pos  int
}

func (s ContinueInstr) InstrString() string {
	if s.pos == 0 {
		return fmt.Sprintf("continue %s", s.loop.stmtname.name)
	}
	return fmt.Sprintf("continue %s (loop is at pos %d)", s.loop.stmtname.name, s.pos)
}

func (s *ContinueInstr) Execute(env *Glisp) error {
	VPrintf("\n executing ContinueInstr with loop: '%#v'\n", s.loop)
	if s.pos == 0 {
		pos, err := env.FindLoop(s.loop)
		if err != nil {
			return err
		}
		s.pos = pos
	}
	env.pc = s.pos + s.loop.continueOffset
	VPrintf("\n  more detail ContinueInstr pos=%d, setting pc = %d\n", s.pos, env.pc)
	return nil
}

type LoopStartInstr struct {
	loop *Loop
}

func (s LoopStartInstr) InstrString() string {
	return fmt.Sprintf("loopstart %s", s.loop.stmtname.name)
}

func (s LoopStartInstr) Execute(env *Glisp) error {
	env.pc++
	return nil
}

// stack cleanup discipline instructions let us
// ensure the stack gets reset to a previous
// known good level. The sym designates
// how far down to clean up, in a unique and
// distinguishable gensym-ed manner.

// create a stack mark
type PushStackmarkInstr struct {
	sym *SexpSymbol
}

func (s PushStackmarkInstr) InstrString() string {
	return fmt.Sprintf("push-stack-mark %s", s.sym.name)
}

func (s PushStackmarkInstr) Execute(env *Glisp) error {
	env.datastack.PushExpr(&SexpStackmark{sym: s.sym})
	env.pc++
	return nil
}

// cleanup until our stackmark, but leave it in place
type PopUntilStackmarkInstr struct {
	sym *SexpSymbol
}

func (s PopUntilStackmarkInstr) InstrString() string {
	return fmt.Sprintf("pop-until-stack-mark %s", s.sym.name)
}

func (s PopUntilStackmarkInstr) Execute(env *Glisp) error {
toploop:
	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		switch m := expr.(type) {
		case *SexpStackmark:
			if m.sym.number == s.sym.number {
				env.datastack.PushExpr(m)
				break toploop
			}
		}
	}
	env.pc++
	return nil
}

// erase everything up-to-and-including our mark
type ClearStackmarkInstr struct {
	sym *SexpSymbol
}

func (s ClearStackmarkInstr) InstrString() string {
	return fmt.Sprintf("clear-stack-mark %s", s.sym.name)
}

func (s ClearStackmarkInstr) Execute(env *Glisp) error {
toploop:
	for {
		expr, err := env.datastack.PopExpr()
		if err != nil {
			return err
		}
		switch m := expr.(type) {
		case *SexpStackmark:
			if m.sym.number == s.sym.number {
				break toploop
			}
		}
	}
	env.pc++
	return nil
}

type DebugInstr struct {
	diagnostic string
}

func (g DebugInstr) InstrString() string {
	return fmt.Sprintf("debug %s", g.diagnostic)
}

func (g DebugInstr) Execute(env *Glisp) error {
	switch g.diagnostic {
	case "showScopes":
		err := env.ShowStackStackAndScopeStack()
		if err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("unknown diagnostic %v", g.diagnostic))
	}
	env.pc++
	return nil
}

// when a defn or fn executes, capture the creation env.
type CreateClosureInstr struct {
	sfun *SexpFunction
}

func (a CreateClosureInstr) InstrString() string {
	return "create closure " + a.sfun.SexpString(0)
}

func (a CreateClosureInstr) Execute(env *Glisp) error {
	env.pc++
	cls := NewClosing(a.sfun.name, env)
	myInvok := a.sfun.Copy()
	myInvok.SetClosing(cls)

	shown, err := myInvok.ShowClosing(env, 8,
		fmt.Sprintf("closedOverScopes of '%s'", myInvok.name))
	if err != nil {
		return err
	}
	VPrintf("+++ CreateClosure: assign to '%s' the stack:\n\n%s\n\n",
		myInvok.SexpString(0), shown)
	top := cls.TopScope()
	VPrintf("222 CreateClosure: top of NewClosing Scope has addr %p and is\n",
		top)
	top.Show(env, 8, fmt.Sprintf("top of NewClosing at %p", top))

	env.datastack.PushExpr(myInvok)
	return nil
}

type AssignInstr struct {
}

func (a AssignInstr) InstrString() string {
	return "assign stack top to stack top -1"
}

func (a AssignInstr) Execute(env *Glisp) error {
	env.pc++
	rhs, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	lhs, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	switch x := lhs.(type) {
	case *SexpSymbol:
		return env.LexicalBindSymbol(x, rhs)
	case Selector:
		Q("AssignInstr: I see lhs is Selector")
		err := x.AssignToSelection(env, rhs)
		return err
	}
	return fmt.Errorf("AssignInstr: don't know how to assign to %T", lhs)
}

func (c *CallInstr) setupDotCallHelper(
	env *Glisp,
	ftop *SexpFunction,
	indirectFuncName *Sexp,
	expressions []Sexp,
	xprBegin int,
	dotSymRef Sexp) {

	Q("\n in CallInstr, fetched out function call from top of datastack.\n")
	*indirectFuncName = ftop
	if ftop.user {
		Q("\n in CallInstr, with user func, passing dot-symbol in directly so assignment will work.\n")
		env.datastack.PushExpr(c.sym)
	} else {
		Q("\n in CallInstr, with sexp func, dereferencing dot-symbol '%s' -> '%s'\n", c.sym.name, dotSymRef.SexpString(0))
		env.datastack.PushExpr(dotSymRef)
	}
	pushme := expressions[xprBegin:]
	for j := range pushme {
		env.datastack.PushExpr(pushme[j])
	}
	Q("\n in CallInstr, after setting up stack for dot-symbol call, datastack:\n")
	//env.datastack.PrintStack()
}
