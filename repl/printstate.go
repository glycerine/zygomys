package zygo

import (
	"reflect"
)

// begin supporting SexpString structs

// PrintState threads the state of display through SexpString() and Show() calls,
// to give pretty-printing indentation and to avoid infinite looping on
// cyclic data structures.
type PrintState struct {
	Indent int
	Seen   Seen
}

func (ps *PrintState) SetSeen(x interface{}) {
	ps.Seen[reflect.ValueOf(x).Pointer()] = struct{}{}
}

func (ps *PrintState) GetSeen(x interface{}) bool {
	_, ok := ps.Seen[reflect.ValueOf(x).Pointer()]
	return ok
}

func (ps *PrintState) GetIndent() int {
	if ps == nil {
		return 0
	}
	return ps.Indent
}

func (ps *PrintState) AddIndent(addme int) *PrintState {
	if ps == nil {
		return &PrintState{
			Indent: addme,
			Seen:   NewSeen(),
		}
	}
	return &PrintState{
		Indent: ps.Indent + addme,
		Seen:   ps.Seen,
	}
}

func NewPrintState() *PrintState {
	return &PrintState{
		Seen: NewSeen(),
	}
}

func NewPrintStateWithIndent(indent int) *PrintState {
	return &PrintState{
		Indent: indent,
		Seen:   NewSeen(),
	}
}

// Seen tracks if a value has already been displayed, to
// detect and avoid cycles
type Seen map[uintptr]struct{}

func NewSeen() Seen {
	return Seen(make(map[uintptr]struct{}))
}

// end supporting SexpString structs
