package glisp

import (
	"fmt"
)

type Instruction interface {
	InstrString() string
}

type JumpInstr struct {
	location int
}

func (j JumpInstr) InstrString() string {
	return fmt.Sprintf("jump %d", j.location)
}

type BranchInstr struct {
	direction bool
	location int
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

type PushInstr struct {
	expr Sexp
}

func (p PushInstr) InstrString() string {
	return "push " + p.expr.SexpString()
}

type PopInstr int

func (p PopInstr) InstrString() string {
	return "pop"
}

type GetInstr struct {
	sym SexpSymbol
}

func (g GetInstr) InstrString() string {
	return fmt.Sprintf("get %s", g.sym.name)
}

type PutInstr struct {
	sym SexpSymbol
}

func (p PutInstr) InstrString() string {
	return fmt.Sprintf("put %s", p.sym.name)
}

type CallInstr struct {
	sym SexpSymbol
	nargs int
}

func (c CallInstr) InstrString() string {
	return fmt.Sprintf("call %s %d", c.sym.name, c.nargs)
}

type DispatchInstr struct {
	nargs int
}

func (d DispatchInstr) InstrString() string {
	return fmt.Sprintf("dispatch %d", d.nargs)
}

type ReturnInstr struct {
	err error
}

func (r ReturnInstr) InstrString() string {
	if r.err == nil {
		return "ret"
	}
	return "ret \"" + r.err.Error() + "\""
}
