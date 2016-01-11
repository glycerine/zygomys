package zygo

import (
	"errors"
	"fmt"
	//"github.com/shurcooL/go-goon"
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

type PushInstrClosure struct {
	expr SexpFunction
}

func (p PushInstrClosure) InstrString() string {
	return "pushC " + p.expr.SexpString()
}

func (p PushInstrClosure) Execute(env *Glisp) error {
	if p.expr.fun != nil {
		p.expr.closeScope = NewStack(ScopeStackSize)

		p.expr.closeScope.PushScope()

		var sym SexpSymbol
		var exp Sexp
		var err error
		for _, v := range p.expr.fun {

			switch it := v.(type) {
			case EnvToStackInstr:
				sym = it.sym
			case PopStackPutEnvInstr:
				sym = it.sym
			case CallInstr:
				sym = it.sym
			default:
				continue
			}

			exp, err, _ = env.scopestack.LookupSymbolNonGlobal(sym)
			if err == nil {
				p.expr.closeScope.BindSymbol(sym, exp)
			}
		}
	} else {
		p.expr.closeScope = env.scopestack.Clone() // for a non script function I have no idea what it accesses, so we clone the whole thing
	}

	env.datastack.PushExpr(p.expr)
	env.pc++
	return nil
}

type PushInstr struct {
	expr Sexp
}

func (p PushInstr) InstrString() string {
	return "push " + p.expr.SexpString()
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
	sym SexpSymbol
}

func (g EnvToStackInstr) InstrString() string {
	return fmt.Sprintf("envToStack %s", g.sym.name)
}

func (g EnvToStackInstr) Execute(env *Glisp) error {

	macxpr, isMacro := env.macros[g.sym.number]
	if isMacro {
		if macxpr.orig != nil {
			return fmt.Errorf("'%s' is a macro, with definition: %s\n", g.sym.name, macxpr.orig.SexpString())
		}
		return fmt.Errorf("'%s' is a builtin macro.\n", g.sym.name)
	}

	expr, err, _ := env.scopestack.LookupSymbol(g.sym)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type PopStackPutEnvInstr struct {
	sym SexpSymbol
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
	return env.scopestack.BindSymbol(p.sym, expr)
}

// Update takes top of datastack and
// assigns it to sym when sym is found
// already in the current scope or
// up the stack. Used
// to implement (set v 10) when v is
// not in the local scope.
type UpdateInstr struct {
	sym SexpSymbol
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

	_, err, scope := env.scopestack.LookupSymbol(p.sym)
	if err != nil {
		// not found up the stack, so treat like (def)
		// instead of (set)
		return env.scopestack.BindSymbol(p.sym, expr)
	}
	// found up the stack, so (set)
	return scope.UpdateSymbolInScope(p.sym, expr)
}

type CallInstr struct {
	sym   SexpSymbol
	nargs int
}

func (c CallInstr) InstrString() string {
	return fmt.Sprintf("call %s %d", c.sym.name, c.nargs)
}

func (c CallInstr) Execute(env *Glisp) error {
	f, ok := env.builtins[c.sym.number]
	if ok {
		return env.CallUserFunction(f, c.sym.name, c.nargs)
	}

	funcobj, err, _ := env.scopestack.LookupSymbol(c.sym)
	if err != nil {
		return err
	}
	switch f := funcobj.(type) {
	case SexpSymbol:
		// allow symbols to refer to functions that we then call
		indirectFuncName, err, _ := env.scopestack.LookupSymbol(f)
		if err != nil {
			return fmt.Errorf("'%s' refers to symbol '%s', but '%s' does not refer to a function.", c.sym.name, f.name, f.name)
		}
		switch g := indirectFuncName.(type) {
		case SexpFunction:
			if !g.user {
				return env.CallFunction(g, c.nargs)
			}
			return env.CallUserFunction(g, f.name, c.nargs)
		default:
			if err != nil {
				return fmt.Errorf("symbol '%s' refers to '%s' which does not refer to a function.", c.sym.name, f.name)
			}
		}

	case SexpFunction:
		if !f.user {
			return env.CallFunction(f, c.nargs)
		}
		return env.CallUserFunction(f, c.sym.name, c.nargs)
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
	case SexpFunction:
		if !f.user {
			return env.CallFunction(f, d.nargs)
		}
		return env.CallUserFunction(f, f.name, d.nargs)
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

type AddScopeInstr int

func (a AddScopeInstr) InstrString() string {
	return "add scope"
}

func (a AddScopeInstr) Execute(env *Glisp) error {
	env.scopestack.PushScope()
	env.pc++
	return nil
}

type RemoveScopeInstr int

func (a RemoveScopeInstr) InstrString() string {
	return "rem scope"
}

func (a RemoveScopeInstr) Execute(env *Glisp) error {
	env.pc++
	return env.scopestack.PopScope()
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
	syms []SexpSymbol
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
		env.scopestack.BindSymbol(bindThisSym, arr[i])
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
	env.datastack.PushExpr(SexpArray(vec))
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
	sym SexpSymbol
}

func (s PushStackmarkInstr) InstrString() string {
	return fmt.Sprintf("push-stack-mark %s", s.sym.name)
}

func (s PushStackmarkInstr) Execute(env *Glisp) error {
	env.datastack.PushExpr(SexpStackmark{sym: s.sym})
	env.pc++
	return nil
}

// cleanup until our stackmark, but leave it in place
type PopUntilStackmarkInstr struct {
	sym SexpSymbol
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
		case SexpStackmark:
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
	sym SexpSymbol
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
		case SexpStackmark:
			if m.sym.number == s.sym.number {
				break toploop
			}
		}
	}
	env.pc++
	return nil
}
