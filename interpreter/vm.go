package glisp

import (
	"errors"
	"fmt"
)

type Instruction interface {
	InstrString() string
	Execute(env *Glisp) error
}

type JumpInstr struct {
	location int
}

var OutOfBounds error = errors.New("jump out of bounds")

func (j JumpInstr) InstrString() string {
	return fmt.Sprintf("jump %d", j.location)
}

func (j JumpInstr) Execute(env *Glisp) error {
	newpc := env.pc + j.location
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
		return JumpInstr{b.location}.Execute(env)
	}
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

type GetInstr struct {
	sym SexpSymbol
}

func (g GetInstr) InstrString() string {
	return fmt.Sprintf("get %s", g.sym.name)
}

func (g GetInstr) Execute(env *Glisp) error {
	expr, err := env.scopestack.LookupSymbol(g.sym)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(expr)
	env.pc++
	return nil
}

type PutInstr struct {
	sym SexpSymbol
}

func (p PutInstr) InstrString() string {
	return fmt.Sprintf("put %s", p.sym.name)
}

func (p PutInstr) Execute(env *Glisp) error {
	expr, err := env.datastack.PopExpr()
	if err != nil {
		return err
	}
	env.pc++
	return env.scopestack.BindSymbol(p.sym, expr)
}

type CallInstr struct {
	sym   SexpSymbol
	nargs int
}

func (c CallInstr) InstrString() string {
	return fmt.Sprintf("call %s %d", c.sym.name, c.nargs)
}

func (c CallInstr) Execute(env *Glisp) error {
	funcobj, err := env.scopestack.LookupSymbol(c.sym)
	if err != nil {
		return err
	}
	switch f := funcobj.(type) {
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
	return errors.New("not a function")
}

type ReturnInstr struct {
	err error
}

func (r ReturnInstr) Execute(env *Glisp) error {
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
